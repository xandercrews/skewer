package network

import (
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"

	dto "github.com/prometheus/client_model/go"
	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/services/base"
	"github.com/stephane-martin/skewer/services/errors"
	"github.com/stephane-martin/skewer/sys/binder"
	"github.com/stephane-martin/skewer/utils"
	"github.com/stephane-martin/skewer/utils/queue/udp"
)

type UdpServerStatus int

const (
	UdpStopped UdpServerStatus = iota
	UdpStarted
)

func initUdpRegistry() {
	base.Once.Do(func() {
		base.InitRegistry()
	})
}

type UdpServiceImpl struct {
	base.BaseService
	UdpConfigs     []conf.UDPSourceConfig
	status         UdpServerStatus
	stasher        base.Stasher
	wg             sync.WaitGroup
	fatalErrorChan chan struct{}
	fatalOnce      *sync.Once

	rawMessagesQueue *udp.Ring
}

func NewUdpService(env *base.ProviderEnv) (base.Provider, error) {
	initUdpRegistry()
	s := UdpServiceImpl{
		status:     UdpStopped,
		stasher:    env.Reporter,
		UdpConfigs: []conf.UDPSourceConfig{},
	}
	s.BaseService.Init()
	s.BaseService.Logger = env.Logger.New("class", "UdpServer")
	s.BaseService.Binder = env.Binder
	return &s, nil
}

func (s *UdpServiceImpl) Type() base.Types {
	return base.UDP
}

//func (s *UdpServiceImpl) SetConf(sc []conf.UDPSourceConfig, pc []conf.ParserConfig, queueSize uint64) {
func (s *UdpServiceImpl) SetConf(c conf.BaseConfig) {
	s.BaseService.Pool = &sync.Pool{New: func() interface{} {
		return &model.RawUdpMessage{}
	}}
	s.BaseService.SetConf(c.Parsers, c.Main.InputQueueSize)
	s.UdpConfigs = c.UDPSource
	s.rawMessagesQueue = udp.NewRing(c.Main.InputQueueSize)
}

// Parse fetch messages from the raw queue, parse them, and push them to be sent.
func (s *UdpServiceImpl) Parse() {
	defer s.wg.Done()

	e := NewParsersEnv(s.ParserConfigs, s.Logger)
	gen := utils.NewGenerator()

	for {
		raw, err := s.rawMessagesQueue.Get()
		if raw == nil || err != nil {
			return
		}
		s.ParseOne(raw, e, gen)
	}
}

func (s *UdpServiceImpl) ParseOne(raw *model.RawUdpMessage, e *ParsersEnv, gen *utils.Generator) {
	defer s.Pool.Put(raw)

	logger := s.Logger.New(
		"protocol", "udp",
		"client", raw.Client,
		"local_port", raw.LocalPort,
		"unix_socket_path", raw.UnixSocketPath,
		"format", raw.Format,
	)

	decoder := utils.SelectDecoder(raw.Encoding)
	parser := e.GetParser(raw.Format)
	if parser == nil {
		logger.Crit("Unknown parser")
		return
	}

	syslogMsg, err := parser.Parse(raw.Message[:raw.Size], decoder, raw.DontParseSD)
	if err != nil {
		base.ParsingErrorCounter.WithLabelValues("udp", raw.Client, raw.Format).Inc()
		//logger.Info("Parsing error", "message", string(raw.Message), "error", err)
		logger.Info("Parsing error", "error", err)
		return
	}
	if syslogMsg == nil {
		return
	}

	fatal, nonfatal := s.stasher.Stash(model.FullMessage{
		Parsed: model.ParsedMessage{
			Fields:         *syslogMsg,
			Client:         raw.Client,
			LocalPort:      raw.LocalPort,
			UnixSocketPath: raw.UnixSocketPath,
		},
		Uid:    gen.Uid(),
		ConfId: raw.ConfID,
	})

	if fatal != nil {
		logger.Error("Fatal error stashing UDP message", "error", fatal)
		s.dofatal()
	} else if nonfatal != nil {
		logger.Warn("Non-fatal error stashing UDP message", "error", nonfatal)
	}
}

func (s *UdpServiceImpl) Gather() ([]*dto.MetricFamily, error) {
	return base.Registry.Gather()
}

func (s *UdpServiceImpl) Start() ([]model.ListenerInfo, error) {
	s.LockStatus()
	defer s.UnlockStatus()
	if s.status != UdpStopped {
		return nil, errors.ServerNotStopped
	}
	s.fatalErrorChan = make(chan struct{})
	s.fatalOnce = &sync.Once{}

	// start the parsers
	cpus := runtime.NumCPU()
	for i := 0; i < cpus; i++ {
		s.wg.Add(1)
		go s.Parse()
	}

	s.ClearConnections()
	infos := s.ListenPacket()
	if len(infos) > 0 {
		s.status = UdpStarted
		s.Logger.Info("Listening on UDP", "nb_services", len(infos))
	} else {
		s.Logger.Debug("The UDP service has not been started: no listening port")
	}
	return infos, nil
}

func (s *UdpServiceImpl) FatalError() chan struct{} {
	return s.fatalErrorChan
}

func (s *UdpServiceImpl) dofatal() {
	s.fatalOnce.Do(func() { close(s.fatalErrorChan) })
}

func (s *UdpServiceImpl) Shutdown() {
	s.Stop()
}

func (s *UdpServiceImpl) Stop() {
	s.LockStatus()
	defer s.UnlockStatus()
	if s.status != UdpStarted {
		return
	}
	s.CloseConnections()
	if s.rawMessagesQueue != nil {
		s.rawMessagesQueue.Dispose()
	}
	s.wg.Wait()

	s.status = UdpStopped
	s.Logger.Debug("Udp server has stopped")
}

func (s *UdpServiceImpl) ListenPacket() []model.ListenerInfo {
	udpinfos := []model.ListenerInfo{}
	s.UnixSocketPaths = []string{}
	for _, syslogConf := range s.UdpConfigs {
		if len(syslogConf.UnixSocketPath) > 0 {
			conn, err := s.Binder.ListenPacket("unixgram", syslogConf.UnixSocketPath)
			if err != nil {
				s.Logger.Warn("Listen unixgram error", "error", err)
			} else {
				s.Logger.Debug(
					"Unixgram listener",
					"protocol", "udp",
					"path", syslogConf.UnixSocketPath,
					"format", syslogConf.Format,
				)
				udpinfos = append(udpinfos, model.ListenerInfo{
					UnixSocketPath: syslogConf.UnixSocketPath,
					Protocol:       "udp",
				})
				s.UnixSocketPaths = append(s.UnixSocketPaths, syslogConf.UnixSocketPath)
				_ = conn.(*binder.FilePacketConn).PacketConn.(*net.UnixConn).SetReadBuffer(65536)
				_ = conn.(*binder.FilePacketConn).PacketConn.(*net.UnixConn).SetWriteBuffer(65536)
				s.wg.Add(1)
				go s.handleConnection(conn, syslogConf)
			}
		} else {
			listenAddrs, _ := syslogConf.GetListenAddrs()
			for port, listenAddr := range listenAddrs {
				conn, err := s.Binder.ListenPacket("udp", listenAddr)
				if err != nil {
					s.Logger.Warn("Listen UDP error", "error", err)
				} else {
					s.Logger.Debug(
						"UDP listener",
						"protocol", "udp",
						"bind_addr", syslogConf.BindAddr,
						"port", port,
						"format", syslogConf.Format,
					)
					udpinfos = append(udpinfos, model.ListenerInfo{
						BindAddr: syslogConf.BindAddr,
						Port:     port,
						Protocol: "udp",
					})
					_ = conn.(*binder.FilePacketConn).PacketConn.(*net.UDPConn).SetReadBuffer(65536)
					_ = conn.(*binder.FilePacketConn).PacketConn.(*net.UDPConn).SetReadBuffer(65536)
					s.wg.Add(1)
					go s.handleConnection(conn, syslogConf)
				}
			}
		}
	}
	return udpinfos
}

func (s *UdpServiceImpl) handleConnection(conn net.PacketConn, config conf.UDPSourceConfig) {
	var localPort int32
	var localPortS string
	var path string
	var err error

	s.AddConnection(conn)

	defer func() {
		s.RemoveConnection(conn)
		s.wg.Done()
	}()

	local := conn.LocalAddr()
	if local != nil {
		l := local.String()
		s := strings.Split(l, ":")
		localPort, err = utils.Atoi32(s[len(s)-1])
		if err != nil {
			path = strings.TrimSpace(l)
		} else {
			localPortS = strconv.FormatInt(int64(localPort), 10)
		}
	}

	logger := s.Logger.New("protocol", "udp", "local_port", localPortS, "unix_socket_path", path)

	// Syslog UDP server
	var remote net.Addr
	var rawmsg *model.RawUdpMessage

	for {
		rawmsg = s.Pool.Get().(*model.RawUdpMessage)
		rawmsg.Size, remote, err = conn.ReadFrom(rawmsg.Message[:])
		if err != nil {
			logger.Debug("Error reading UDP", "error", err)
			s.Pool.Put(rawmsg)
			return
		}
		if rawmsg.Size == 0 {
			s.Pool.Put(rawmsg)
			continue
		}
		rawmsg.LocalPort = localPort
		rawmsg.UnixSocketPath = path
		rawmsg.Encoding = config.Encoding
		rawmsg.Format = config.Format
		rawmsg.ConfID = config.ConfID
		rawmsg.DontParseSD = config.DontParseSD
		rawmsg.Client = ""
		if remote == nil {
			// unix socket
			rawmsg.Client = "localhost"
		} else {
			rawmsg.Client = strings.Split(remote.String(), ":")[0]
		}

		err := s.rawMessagesQueue.Put(rawmsg)
		if err != nil {
			logger.Warn("Error queueing UDP message", "error", err)
			return
		}
		base.IncomingMsgsCounter.WithLabelValues("udp", rawmsg.Client, localPortS, path).Inc()
	}
}
