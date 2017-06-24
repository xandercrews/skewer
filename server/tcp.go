package server

import (
	"bufio"
	"bytes"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/oklog/ulid"
	"github.com/stephane-martin/relp2kafka/conf"
	"github.com/stephane-martin/relp2kafka/model"
	"github.com/stephane-martin/relp2kafka/store"
)

type TcpServerStatus int

const (
	TcpStopped TcpServerStatus = iota
	TcpStarted
)

type TcpServer struct {
	StreamServer
	status     TcpServerStatus
	ClosedChan chan TcpServerStatus
	store      *store.MessageStore
}

func (s *TcpServer) init() {
	s.StreamServer.init()
}

func NewTcpServer(c *conf.GConfig, st *store.MessageStore, logger log15.Logger) *TcpServer {
	s := TcpServer{}
	s.logger = logger.New("class", "TcpServer")
	s.init()
	s.protocol = "tcp"
	s.Conf = *c
	s.handler = TcpHandler{Server: &s}
	s.status = TcpStopped
	s.store = st

	return &s
}

func (s *TcpServer) Start() (err error) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()
	if s.status != TcpStopped {
		err = ServerNotStopped
		return
	}
	s.ClosedChan = make(chan TcpServerStatus, 1)

	s.initParsers()
	// start listening on the required ports
	nb := s.initTCPListeners()
	if nb > 0 {
		s.status = TcpStarted
		s.Listen()
	} else {
		s.logger.Info("TCP Server not started: no listening port")
		close(s.ClosedChan)
	}
	return

}

func (s *TcpServer) Stop() {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()
	if s.status != TcpStarted {
		return
	}
	s.resetTCPListeners() // close the listeners. This will make Listen to return and close all current connections.
	s.wg.Wait()           // wait that all HandleConnection goroutines have ended
	s.logger.Debug("TcpServer goroutines have ended")

	s.status = TcpStopped
	s.ClosedChan <- TcpStopped
	close(s.ClosedChan)
	s.logger.Info("TCP server has stopped")
}

type TcpHandler struct {
	Server *TcpServer
}

func (h TcpHandler) HandleConnection(conn net.Conn, i int) {
	s := h.Server
	s.AddConnection(conn)

	raw_messages_chan := make(chan *model.RawMessage)

	defer func() {
		close(raw_messages_chan)
		s.RemoveConnection(conn)
		s.wg.Done()
	}()

	var client string
	remote := conn.RemoteAddr()
	if remote != nil {
		client = strings.Split(remote.String(), ":")[0]
	}

	var local_port int
	local := conn.LocalAddr()
	if local != nil {
		s := strings.Split(local.String(), ":")
		local_port, _ = strconv.Atoi(s[len(s)-1])
	}

	logger := s.logger.New("remote", client, "local_port", local_port)
	logger.Info("New TCP client")

	// pull messages from raw_messages_chan, parse them and push them to the Store
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
		for m := range raw_messages_chan {
			parser := s.GetParser(s.Conf.Syslog[i].Format)
			if parser == nil {
				// todo: log
				continue
			}
			p, err := parser.Parse(m.Message, s.Conf.Syslog[i].DontParseSD)

			if err == nil {
				uid, err := ulid.New(ulid.Timestamp(p.TimeReported), entropy)
				if err != nil {
					// should not happen
					s.logger.Error("Error generating a ULID", "error", err)
				} else {
					parsed_msg := model.TcpUdpParsedMessage{
						Parsed: model.ParsedMessage{
							Fields:    p,
							Client:    m.Client,
							LocalPort: m.LocalPort,
						},
						Uid:       uid.String(),
						ConfIndex: i,
					}
					s.store.Inputs <- &parsed_msg
				}
			} else {
				logger.Info("Parsing error", "Message", m.Message, "error", err)
			}
		}
	}()

	timeout := s.Conf.Syslog[i].Timeout
	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	}
	scanner := bufio.NewScanner(conn)
	switch s.Conf.Syslog[i].Format {
	case "rfc5424", "rfc3164", "json", "auto":
		scanner.Split(TcpSplit)
	default:
		scanner.Split(LFTcpSplit)
	}

	for {
		if scanner.Scan() {
			if timeout > 0 {
				conn.SetReadDeadline(time.Now().Add(timeout))
			}
			raw := model.RawMessage{
				Client:    client,
				LocalPort: local_port,
				Message:   scanner.Text(),
			}
			raw_messages_chan <- &raw
		} else {
			logger.Info("Scanning the TCP stream has ended", "error", scanner.Err())
			return
		}
	}
}

func LFTcpSplit(data []byte, atEOF bool) (int, []byte, error) {
	trimmed_data := bytes.TrimLeft(data, " \r\n")
	if len(trimmed_data) == 0 {
		return 0, nil, nil
	}
	trimmed := len(data) - len(trimmed_data)
	lf := bytes.IndexByte(trimmed_data, '\n')
	if lf >= 0 {
		token := bytes.Trim(trimmed_data[0:lf], " \r\n")
		advance := trimmed + lf + 1
		return advance, token, nil
	} else {
		// data does not contain a full syslog line
		return 0, nil, nil
	}
}

func TcpSplit(data []byte, atEOF bool) (int, []byte, error) {
	trimmed_data := bytes.TrimLeft(data, " \r\n")
	if len(trimmed_data) == 0 {
		return 0, nil, nil
	}
	trimmed := len(data) - len(trimmed_data)
	if trimmed_data[0] == byte('<') {
		// LF framing
		lf := bytes.IndexByte(trimmed_data, '\n')
		if lf >= 0 {
			token := bytes.Trim(trimmed_data[0:lf], " \r\n")
			advance := trimmed + lf + 1
			return advance, token, nil
		} else {
			// data does not contain a full syslog line
			return 0, nil, nil
		}
	} else {
		// octet counting framing
		sp := bytes.IndexAny(trimmed_data, " \n")
		if sp <= 0 {
			return 0, nil, nil
		}
		datalen_s := bytes.Trim(trimmed_data[0:sp], " \r\n")
		datalen, err := strconv.Atoi(string(datalen_s))
		if err != nil {
			return 0, nil, err
		}
		advance := trimmed + sp + 1 + datalen
		if len(data) >= advance {
			token := bytes.Trim(trimmed_data[sp+1:sp+1+datalen], " \r\n")
			return advance, token, nil
		} else {
			return 0, nil, nil
		}

	}
}
