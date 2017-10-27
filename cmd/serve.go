package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/kardianos/osext"
	"github.com/spf13/cobra"
	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/consul"
	"github.com/stephane-martin/skewer/journald"
	"github.com/stephane-martin/skewer/metrics"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/services"
	"github.com/stephane-martin/skewer/sys"
	"github.com/stephane-martin/skewer/sys/binder"
	"github.com/stephane-martin/skewer/sys/capabilities"
	"github.com/stephane-martin/skewer/sys/dumpable"
	"github.com/stephane-martin/skewer/utils/logging"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start listening for Syslog messages and forward them to Kafka",
	Long: `The serve command is the main skewer command. It launches a long
running process that listens to syslog messages according to the configuration,
connects to Kafka, and forwards messages to Kafka.`,
	Run: func(cmd *cobra.Command, args []string) {
		runserve()
	},
}

type spair struct {
	child  int
	parent int
}

var testFlag bool
var syslogFlag bool
var loglevelFlag string
var logfilenameFlag string
var logjsonFlag bool
var pidFilenameFlag string
var consulRegisterFlag bool
var consulServiceName string
var uidFlag string
var gidFlag string
var dumpableFlag bool
var profile bool
var handles []string
var handlesMap map[string]int

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().BoolVar(&testFlag, "test", false, "Print messages to stdout instead of sending to Kafka")
	serveCmd.Flags().BoolVar(&syslogFlag, "syslog", false, "Send logs to the local syslog (are you sure you wan't to do that ?)")
	serveCmd.Flags().StringVar(&loglevelFlag, "loglevel", "info", "Set logging level")
	serveCmd.Flags().StringVar(&logfilenameFlag, "logfilename", "", "Write logs to a file instead of stderr")
	serveCmd.Flags().BoolVar(&logjsonFlag, "logjson", false, "Write logs in JSON format")
	serveCmd.Flags().StringVar(&pidFilenameFlag, "pidfile", "", "If given, write PID to file")
	serveCmd.Flags().BoolVar(&consulRegisterFlag, "register", false, "Register services in consul")
	serveCmd.Flags().StringVar(&consulServiceName, "servicename", "skewer", "Service name to register in consul")
	serveCmd.Flags().StringVar(&uidFlag, "uid", "", "Switch to this user ID (when launched as root)")
	serveCmd.Flags().StringVar(&gidFlag, "gid", "", "Switch to this group ID (when launched as root)")
	serveCmd.Flags().BoolVar(&dumpableFlag, "dumpable", false, "if set, the skewer process will be traceable/dumpable")
	serveCmd.Flags().BoolVar(&profile, "profile", false, "if set, profile memory")

	handles = []string{
		"CHILD_BINDER",
		"TCP_BINDER",
		"UDP_BINDER",
		"RELP_BINDER",
		"CHILD_LOGGER",
		"TCP_LOGGER",
		"UDP_LOGGER",
		"RELP_LOGGER",
		"JOURNAL_LOGGER",
		"CONFIG_LOGGER",
		"STORE_LOGGER",
		"ACCT_LOGGER",
	}

	handlesMap = map[string]int{}
	for i, h := range handles {
		handlesMap[h] = i + 3
	}
}

func runserve() {

	if !dumpableFlag {
		err := dumpable.SetNonDumpable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting PR_SET_DUMPABLE: %s\n", err)
		}
	}

	if os.Getenv("SKEWER_LINUX_CHILD") == "TRUE" {
		// we are in the final child on linux
		err := capabilities.NoNewPriv()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		err = Serve()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Fatal error in Serve():", err)
			os.Exit(-1)
		}
		os.Exit(0)
	}

	if os.Getenv("SKEWER_CHILD") == "TRUE" {
		// we are in the child
		if capabilities.CapabilitiesSupported {
			// another execve is necessary on Linux to ensure that
			// the following capability drop will be effective on
			// all go threads
			runtime.LockOSThread()
			err := capabilities.DropNetBind()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			exe, err := osext.Executable()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			err = syscall.Exec(exe, os.Args, []string{"PATH=/bin:/usr/bin", "SKEWER_LINUX_CHILD=TRUE"})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		} else {
			err := capabilities.NoNewPriv()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			err = Serve()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Fatal error in Serve()", err)
				os.Exit(-1)
			}
			os.Exit(0)
		}
	}

	// we are in the parent
	if capabilities.CapabilitiesSupported {
		// under Linux, re-exec ourself immediately with fewer privileges
		runtime.LockOSThread()
		need_fix, err := capabilities.NeedFixLinuxPrivileges(uidFlag, gidFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		if need_fix {
			if os.Getenv("SKEWER_DROPPED") == "TRUE" {
				fmt.Fprintln(os.Stderr, "Dropping privileges failed!")
				fmt.Fprintln(os.Stderr, "Uid", os.Getuid())
				fmt.Fprintln(os.Stderr, "Gid", os.Getgid())
				fmt.Fprintln(os.Stderr, "Capabilities")
				fmt.Fprintln(os.Stderr, capabilities.GetCaps())
				os.Exit(-1)
			}
			err = capabilities.FixLinuxPrivileges(uidFlag, gidFlag)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			err = capabilities.NoNewPriv()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			exe, err := os.Executable()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			err = syscall.Exec(exe, os.Args, []string{"PATH=/bin:/usr/bin", "SKEWER_DROPPED=TRUE"})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		}
	}

	rootlogger := logging.SetLogging(loglevelFlag, logjsonFlag, syslogFlag, logfilenameFlag)
	logger := rootlogger.New("proc", "parent")

	mustSocketPair := func(typ int) spair {
		a, b, err := sys.SocketPair(typ)
		if err != nil {
			logger.Crit("SocketPair() error", "error", err)
			os.Exit(-1)
		}
		return spair{child: a, parent: b}
	}

	getLoggerConn := func(handle int) net.Conn {
		loggerConn, _ := net.FileConn(os.NewFile(uintptr(handle), "logger"))
		loggerConn.(*net.UnixConn).SetReadBuffer(65536)
		loggerConn.(*net.UnixConn).SetWriteBuffer(65536)
		return loggerConn
	}

	numuid, numgid, err := sys.LookupUid(uidFlag, gidFlag)
	if err != nil {
		logger.Crit("Error looking up uid", "error", err, "uid", uidFlag, "gid", gidFlag)
		os.Exit(-1)
	}
	if numuid == 0 {
		logger.Crit("Provide a non-privileged user with --uid flag")
		os.Exit(-1)
	}

	binderSockets := map[string]spair{}
	loggerSockets := map[string]spair{}

	for _, h := range handles {
		if strings.HasSuffix(h, "_BINDER") {
			binderSockets[h] = mustSocketPair(syscall.SOCK_STREAM)
		} else {
			loggerSockets[h] = mustSocketPair(syscall.SOCK_DGRAM)
		}
	}

	binderParents := []int{}
	for _, s := range binderSockets {
		binderParents = append(binderParents, s.parent)
	}
	err = binder.Binder(binderParents, logger) // returns immediately
	if err != nil {
		logger.Crit("Error setting the root binder", "error", err)
		os.Exit(-1)
	}

	remoteLoggerConn := []net.Conn{}
	for _, s := range loggerSockets {
		remoteLoggerConn = append(remoteLoggerConn, getLoggerConn(s.parent))
	}
	logging.LogReceiver(context.Background(), rootlogger, remoteLoggerConn)

	logger.Debug("Target user", "uid", numuid, "gid", numgid)

	// execute child under the new user
	exe, err := osext.Executable() // custom Executable function to support OpenBSD
	if err != nil {
		logger.Crit("Error getting executable name", "error", err)
		os.Exit(-1)
	}

	extraFiles := []*os.File{}
	for _, h := range handles {
		if strings.HasSuffix(h, "_BINDER") {
			extraFiles = append(extraFiles, os.NewFile(uintptr(binderSockets[h].child), h))
		} else {
			extraFiles = append(extraFiles, os.NewFile(uintptr(loggerSockets[h].child), h))
		}
	}

	childProcess := exec.Cmd{
		Args:       os.Args,
		Path:       exe,
		Stdin:      nil,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		ExtraFiles: extraFiles,
		Env:        []string{"SKEWER_CHILD=TRUE", "PATH=/bin:/usr/bin"},
	}
	if os.Getuid() != numuid {
		childProcess.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uint32(numuid), Gid: uint32(numgid)}}
	}
	err = childProcess.Start()
	if err != nil {
		logger.Crit("Error starting child", "error", err)
		os.Exit(-1)
	}

	for _, h := range handles {
		if strings.HasSuffix(h, "_BINDER") {
			syscall.Close(binderSockets[h].child)
		} else {
			syscall.Close(loggerSockets[h].child)
		}
	}

	sig_chan := make(chan os.Signal, 10)
	once := sync.Once{}
	go func() {
		for sig := range sig_chan {
			logger.Debug("parent received signal", "signal", sig)
			if sig == syscall.SIGTERM {
				once.Do(func() { childProcess.Process.Signal(sig) })
			} else if sig == syscall.SIGHUP {
				childProcess.Process.Signal(sig)
			}
		}
	}()
	signal.Notify(sig_chan, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	logger.Debug("PIDs", "parent", os.Getpid(), "child", childProcess.Process.Pid)

	childProcess.Process.Wait()
	os.Exit(0)

}

func Serve() error {
	globalCtx, gCancel := context.WithCancel(context.Background())
	defer gCancel()
	shutdownCtx, shutdown := context.WithCancel(globalCtx)
	defer shutdown()
	var logger log15.Logger

	binderFile := os.NewFile(uintptr(handlesMap["CHILD_BINDER"]), "binder")

	loggerConn, _ := net.FileConn(os.NewFile(uintptr(handlesMap["CHILD_LOGGER"]), "logger"))
	loggerConn.(*net.UnixConn).SetReadBuffer(65536)
	loggerConn.(*net.UnixConn).SetWriteBuffer(65536)
	loggerCtx, cancelLogger := context.WithCancel(context.Background())
	defer cancelLogger()
	logger = logging.NewRemoteLogger(loggerCtx, loggerConn).New("proc", "child")

	logger.Debug("Serve() runs under user", "uid", os.Getuid(), "gid", os.Getgid())
	if capabilities.CapabilitiesSupported {
		logger.Debug("Capabilities", "caps", capabilities.GetCaps())
	}

	binderClient, err := binder.NewBinderClient(binderFile, logger)
	if err != nil {
		logger.Error("Error binding to the root parent socket", "error", err)
		binderClient = nil
	} else {
		defer binderClient.Quit()
	}

	params := consul.ConnParams{
		Address:    consulAddr,
		Datacenter: consulDC,
		Token:      consulToken,
		CAFile:     consulCAFile,
		CAPath:     consulCAPath,
		CertFile:   consulCertFile,
		KeyFile:    consulKeyFile,
		Insecure:   consulInsecure,
		Prefix:     consulPrefix,
	}

	confSvc := services.NewConfigurationService(handlesMap["CONFIG_LOGGER"], logger)
	startConfSvc := func() chan *conf.BaseConfig {
		confSvc.SetConfDir(configDirName)
		confSvc.SetConsulParams(params)
		err = confSvc.Start()
		if err != nil {
			logger.Error("Error starting the configuration service", "error", err)
			return nil
		}
		return confSvc.Chan()
	}

	newConfChannel := startConfSvc()
	if newConfChannel == nil {
		time.Sleep(200 * time.Millisecond)
		return fmt.Errorf("Error starting the configuration service")
	}

	c := <-newConfChannel
	c.Store.Dirname = storeDirname
	logger.Info("Store location", "path", c.Store.Dirname)

	// create a consul consulRegistry
	var consulRegistry *consul.Registry
	if consulRegisterFlag {
		consulRegistry, err = consul.NewRegistry(globalCtx, params, consulServiceName, logger)
		if err != nil {
			consulRegistry = nil
		}
	}

	metricsServer := &metrics.MetricsServer{}

	// setup the Store
	st := services.NewStorePlugin(handlesMap["STORE_LOGGER"], logger)
	st.SetConf(*c)
	err = st.Create(testFlag, dumpableFlag, storeDirname, "", "")
	if err != nil {
		logger.Crit("Can't create the message Store", "error", err)
		time.Sleep(100 * time.Millisecond)
		return err
	}
	_, err = st.Start()
	if err != nil {
		logger.Crit("Can't start the forwarder", "error", err)
		time.Sleep(100 * time.Millisecond)
		return err
	}

	sig_chan := make(chan os.Signal, 10)
	signal.Notify(sig_chan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	relpServicePlugin := services.NewPluginController(
		services.RELP, st, consulRegistry,
		handlesMap["RELP_BINDER"], handlesMap["RELP_LOGGER"],
		logger,
	)

	tcpServicePlugin := services.NewPluginController(
		services.TCP, st, consulRegistry,
		handlesMap["TCP_BINDER"], handlesMap["TCP_LOGGER"],
		logger,
	)

	udpServicePlugin := services.NewPluginController(
		services.UDP, st, consulRegistry,
		handlesMap["UDP_BINDER"], handlesMap["UDP_LOGGER"],
		logger,
	)

	journalServicePlugin := services.NewPluginController(
		services.Journal, st, consulRegistry,
		0, handlesMap["JOURNAL_LOGGER"],
		logger,
	)
	accountingServicePlugin := services.NewPluginController(
		services.Accounting, st, consulRegistry,
		0, handlesMap["ACCT_LOGGER"],
		logger,
	)

	startAccounting := func(curconf *conf.BaseConfig) {
		if curconf.Accounting.Enabled {
			logger.Info("Process accounting is enabled")
			err := accountingServicePlugin.Create(testFlag, dumpableFlag, "", "", curconf.Accounting.Path)
			if err != nil {
				logger.Warn("Error creating the accounting plugin", "error", err)
				return
			}
			accountingServicePlugin.SetConf(*curconf)
			_, err = accountingServicePlugin.Start()
			if err != nil {
				logger.Warn("Error starting accounting plugin", "error", err)
			} else {
				logger.Debug("Accounting plugin has been started")
			}
		}
	}

	startJournal := func(curconf *conf.BaseConfig) {
		// retrieve messages from journald
		if journald.Supported {
			logger.Info("Journald is supported")
			if curconf.Journald.Enabled {
				logger.Info("Journald service is enabled")
				// in fact Create() will only do something the first time startJournal() is called
				err := journalServicePlugin.Create(testFlag, dumpableFlag, "", "", "")
				if err != nil {
					logger.Warn("Error creating Journald plugin", "error", err)
					return
				}
				journalServicePlugin.SetConf(*curconf)
				_, err = journalServicePlugin.Start()
				if err != nil {
					logger.Error("Error starting Journald plugin", "error", err)
				} else {
					logger.Debug("Journald plugin has been started")
				}
			} else {
				logger.Info("Journald service is disabled")
			}
		} else {
			logger.Info("Journald service is not supported (only Linux)")
		}
	}

	startRELP := func(curconf *conf.BaseConfig) {
		err := relpServicePlugin.Create(testFlag, dumpableFlag, "", "", "")
		if err != nil {
			logger.Warn("Error creating RELP plugin", "error", err)
			return
		}
		relpServicePlugin.SetConf(*curconf)
		_, err = relpServicePlugin.Start()
		if err != nil {
			logger.Warn("Error starting RELP plugin", "error", err)
		} else {
			logger.Debug("RELP plugin has been started")
		}
	}

	var tcpinfos []model.ListenerInfo

	startTCP := func(curconf *conf.BaseConfig) {
		err := tcpServicePlugin.Create(testFlag, dumpableFlag, "", "", "")
		if err != nil {
			logger.Warn("Error creating TCP plugin", "error", err)
			return
		}
		tcpServicePlugin.SetConf(*curconf)
		tcpinfos, err = tcpServicePlugin.Start()
		if err != nil {
			logger.Warn("Error starting TCP plugin", "error", err)
		} else if len(tcpinfos) == 0 {
			logger.Info("TCP plugin not started")
		} else {
			logger.Debug("TCP plugin has been started", "listeners", len(tcpinfos))
		}
	}

	startUDP := func(curconf *conf.BaseConfig) {
		err := udpServicePlugin.Create(testFlag, dumpableFlag, "", "", "")
		if err != nil {
			logger.Warn("Error creating UDP plugin", "error", err)
			return
		}
		udpServicePlugin.SetConf(*curconf)
		udpinfos, err := udpServicePlugin.Start()
		if err != nil {
			logger.Warn("Error starting UDP plugin", "error", err)
		} else if len(udpinfos) == 0 {
			logger.Info("UDP plugin not started")
		} else {
			logger.Debug("UDP plugin started", "listeners", len(udpinfos))
		}
	}

	stopTCP := func() {
		tcpServicePlugin.Shutdown(3 * time.Second)
	}

	stopUDP := func() {
		udpServicePlugin.Shutdown(3 * time.Second)
	}

	stopRELP := func() {
		relpServicePlugin.Shutdown(10 * time.Second)
	}

	stopAccounting := func() {
		accountingServicePlugin.Shutdown(3 * time.Second)
	}

	stopJournal := func(sht bool) {
		if journald.Supported {
			if sht {
				journalServicePlugin.Shutdown(5 * time.Second)
			} else {
				// we keep the same instance of the journald plugin, so
				// that we can continue to fetch messages from a
				// consistent position in journald
				journalServicePlugin.Stop()
			}
		}
	}

	Reload := func(newConf *conf.BaseConfig) (fatal error) {
		logger.Info("Reloading configuration")
		// first, let's stop the HTTP server that reports the metrics
		metricsServer.Stop()
		// stop the kafka forwarder
		st.Stop()
		logger.Debug("The forwarder has been stopped")
		st.SetConf(*newConf)
		// restart the kafka forwarder
		_, fatal = st.Start()
		if fatal != nil {
			return fatal
		}

		wg := &sync.WaitGroup{}

		if journald.Supported {
			// restart the journal service
			wg.Add(1)
			go func() {
				stopJournal(false)
				startJournal(newConf)
				wg.Done()
			}()
		}

		// restart the accounting service
		wg.Add(1)
		go func() {
			stopAccounting()
			startAccounting(newConf)
			wg.Done()
		}()

		// restart the RELP service
		wg.Add(1)
		go func() {
			stopRELP()
			startRELP(newConf)
			wg.Done()
		}()

		// restart the TCP service
		wg.Add(1)
		go func() {
			stopTCP()
			startTCP(newConf)
			wg.Done()
		}()

		// restart the UDP service
		wg.Add(1)
		go func() {
			stopUDP()
			startUDP(newConf)
			wg.Done()
		}()
		wg.Wait()

		// restart the HTTP metrics server
		metricsServer.NewConf(
			newConf.Metrics,
			journalServicePlugin,
			accountingServicePlugin,
			relpServicePlugin,
			tcpServicePlugin,
			udpServicePlugin,
			st,
		)
		return nil
	}

	destructor := func() {
		stopRELP()
		logger.Debug("The RELP service has been stopped")

		stopAccounting()
		logger.Debug("Stopped accounting service")

		stopJournal(true)
		logger.Debug("Stopped journald service")

		stopTCP()
		logger.Debug("The TCP service has been stopped")

		stopUDP()
		logger.Debug("The UDP service has been stopped")

		confSvc.Stop()

		gCancel()
		st.Shutdown(5 * time.Second)
		if consulRegistry != nil {
			consulRegistry.WaitFinished() // wait that the services have been unregistered from Consul
		}
		cancelLogger()
		time.Sleep(time.Second)
	}

	defer destructor()

	startJournal(c)
	startAccounting(c)
	startRELP(c)
	startTCP(c)
	startUDP(c)

	metricsServer.NewConf(
		c.Metrics,
		journalServicePlugin,
		accountingServicePlugin,
		relpServicePlugin,
		tcpServicePlugin,
		udpServicePlugin,
		st,
	)

	if profile {
		go func() {
			mux := http.NewServeMux()
			mux.Handle("/pprof/heap", pprof.Handler("heap"))
			mux.Handle("/pprof/profile", http.HandlerFunc(pprof.Profile))
			server := &http.Server{
				Addr:    "127.0.0.1:6600",
				Handler: mux,
			}
			server.ListenAndServe()
		}()
	}

	logger.Debug("Main loop is starting")
	for {
		select {
		case <-shutdownCtx.Done():
			logger.Info("Shutting down")
			return nil
		default:
		}

		select {
		case <-st.ShutdownChan:
			logger.Crit("Abnormal shutdown of the Store: aborting all operations")
			shutdown()
		default:
		}

		select {
		case <-shutdownCtx.Done():
		case <-st.ShutdownChan:

		case newConf, more := <-newConfChannel:
			if more {
				newConf.Store = c.Store
				c = newConf
				err := Reload(c)
				if err != nil {
					logger.Crit("Fatal error when reloading configuration", "error", err)
					shutdown()
				}
			} else {
				// newConfChannel has been closed ?!
				select {
				case <-shutdownCtx.Done():
					// this is normal, we are shutting down
				default:
					// not normal, let's try to restart the service
					newConfChannel = startConfSvc()
					if newConfChannel == nil {
						logger.Crit("Can't restart the configuration service: aborting all operations")
						shutdown()
					} else {
						logger.Warn("Configuration service has been restarted")
					}
				}
			}

		case sig := <-sig_chan:
			if sig == syscall.SIGHUP {
				signal.Stop(sig_chan)
				signal.Ignore(syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
				select {
				case <-shutdownCtx.Done():
				default:
					logger.Info("SIGHUP received: reloading configuration")
					confSvc.Reload()
					sig_chan = make(chan os.Signal, 10)
					signal.Notify(sig_chan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
				}

			} else if sig == syscall.SIGTERM || sig == syscall.SIGINT {
				signal.Stop(sig_chan)
				signal.Ignore(syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
				sig_chan = nil
				logger.Info("Termination signal received", "signal", sig)
				shutdown()
			} else {
				logger.Warn("Unknown signal received", "signal", sig)
			}

		}
	}
}
