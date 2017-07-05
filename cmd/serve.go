package cmd

import (
	"context"
	"fmt"
	"log/syslog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
	"github.com/stephane-martin/relp2kafka/conf"
	"github.com/stephane-martin/relp2kafka/consul"
	"github.com/stephane-martin/relp2kafka/metrics"
	"github.com/stephane-martin/relp2kafka/server"
	"github.com/stephane-martin/relp2kafka/store"
	"github.com/stephane-martin/relp2kafka/sys"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		Serve()
	},
}

var testFlag bool
var syslogFlag bool
var loglevelFlag string
var logfilenameFlag string
var logjsonFlag bool
var pidFilenameFlag string

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().BoolVar(&testFlag, "test", false, "Print to stdout instead of sending to Kafka")
	serveCmd.Flags().BoolVar(&syslogFlag, "syslog", false, "Send logs to the local syslog")
	serveCmd.Flags().StringVar(&loglevelFlag, "loglevel", "info", "Set logging level")
	serveCmd.Flags().StringVar(&logfilenameFlag, "logfilename", "", "Write logs to a file instead of stderr")
	serveCmd.Flags().BoolVar(&logjsonFlag, "json", false, "Write logs in JSON format")
	serveCmd.Flags().StringVar(&pidFilenameFlag, "pidfile", "", "If given, write PID to file")

}

func SetLogging() log15.Logger {
	logger := log15.New()
	log_handlers := []log15.Handler{}
	var formatter log15.Format
	if logjsonFlag {
		formatter = log15.JsonFormat()
	} else {
		formatter = log15.LogfmtFormat()
	}
	if syslogFlag {
		h, e := log15.SyslogHandler(syslog.LOG_LOCAL0|syslog.LOG_DEBUG, "relp2kafka", formatter)
		if e != nil {
			fmt.Printf("Error opening syslog file: %s\n", e)
			os.Exit(-1)
		}
		log_handlers = append(log_handlers, h)
	}
	logfilenameFlag = strings.TrimSpace(logfilenameFlag)
	if len(logfilenameFlag) > 0 {
		h, e := log15.FileHandler(logfilenameFlag, formatter)
		if e != nil {
			fmt.Printf("Error opening log file '%s': %s\n", logfilenameFlag, e)
			os.Exit(-1)
		}
		log_handlers = append(log_handlers, h)
	}
	if len(log_handlers) == 0 {
		log_handlers = []log15.Handler{log15.StderrHandler}
	}
	handler := log15.MultiHandler(log_handlers...)

	lvl, e := log15.LvlFromString(loglevelFlag)
	if e != nil {
		lvl = log15.LvlInfo
	}
	handler = log15.LvlFilterHandler(lvl, handler)

	logger.SetHandler(handler)
	return logger
}

func Serve() {
	gctx, gCancel := context.WithCancel(context.Background())
	shutdownCtx, shutdown := context.WithCancel(gctx)
	watchCtx, stopWatch := context.WithCancel(shutdownCtx)

	defer func() {
		gCancel()
		time.Sleep(time.Second)
	}()

	logger := SetLogging()

	var err error
	err = sys.SetNonDumpable()
	if err != nil {
		logger.Warn("Error setting PR_SET_DUMPABLE", "error", err)
	}

	var c *conf.GConfig
	var st *store.MessageStore
	var updated chan bool
	params := consul.ConnParams{Address: consulAddr, Datacenter: consulDC, Token: consulToken}

	// read configuration
	for {
		c, updated, err = conf.InitLoad(watchCtx, configDirName, params, consulPrefix, logger)
		if err == nil {
			break
		}
		logger.Error("Error getting configuration. Sleep and retry.", "error", err)
		time.Sleep(30 * time.Second)
	}

	// prepare the message store
	st, err = store.NewStore(gctx, c.Store.Dirname, c.Store.Maxsize, c.Store.FSync, logger, testFlag)
	if err != nil {
		logger.Crit("Can't create the message Store", "error", err)
		os.Exit(-1)
	}
	st.SendToKafka(c.Kafka)

	metrics := metrics.SetupMetrics()

	// retrieve messages from journald
	var journaldServer *server.JournaldServer
	if c.Journald.Enabled {
		logger.Info("Journald is enabled")
		journaldServer, err = server.NewJournaldServer(gctx, c.Journald, st, metrics, logger)
		if err == nil {
			journaldServer.Start()
		} else {
			// todo: log
		}
	}

	// prepare the RELP service
	relpServer := server.NewRelpServer(c, metrics, logger)
	if testFlag {
		relpServer.SetTest()
	}
	sig_chan := make(chan os.Signal)
	signal.Notify(sig_chan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	relpServer.StatusChan <- server.Stopped // trigger the RELP service to start

	// start the TCP service
	tcpServer := server.NewTcpServer(c, st, metrics, logger)
	if testFlag {
		tcpServer.SetTest()
	}
	err = tcpServer.Start()
	if err != nil {
		logger.Error("Error starting the TCP server", "error", err)
	}

	// start the UDP service
	udpServer := server.NewUdpServer(c, st, metrics, logger)
	if testFlag {
		udpServer.SetTest()
	}
	err = udpServer.Start()
	if err != nil {
		logger.Error("Error starting the UDP server", "error", err)
	}

	Reload := func(newConf *conf.GConfig) {
		st.StopSendToKafka()
		st.SendToKafka(newConf.Kafka)
		wg := &sync.WaitGroup{}

		if journaldServer != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				journaldServer.Stop()
				journaldServer.Conf = newConf.Journald
				journaldServer.Start()
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			relpServer.Stop()
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			tcpServer.Stop()
			<-tcpServer.ClosedChan
			tcpServer.Conf = *newConf
			err := tcpServer.Start()
			if err != nil {
				logger.Error("Error starting the TCP server", "error", err)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			udpServer.Stop()
			<-udpServer.ClosedChan
			udpServer.Conf = *newConf
			err := udpServer.Start()
			if err != nil {
				logger.Error("Error starting the UDP server", "error", err)
			}
		}()
		wg.Wait()
	}

	for {
		select {
		case <-shutdownCtx.Done():
			logger.Info("Shutting down")
			relpServer.FinalStop()
			tcpServer.Stop()
			udpServer.Stop()
			<-tcpServer.ClosedChan
			logger.Info("The TCP service has been stopped")
			<-udpServer.ClosedChan
			logger.Info("The UDP service has been stopped")
			return

		case _, more := <-updated:
			if more {
				select {
				case <-shutdownCtx.Done():
				default:
					logger.Info("Configuration was updated by Consul")
					Reload(c)
				}
			}

		case sig := <-sig_chan:
			if sig == syscall.SIGHUP {
				select {
				case <-shutdownCtx.Done():
				default:
					logger.Info("SIGHUP received: reloading configuration")
					newWatchCtx, newStopWatch := context.WithCancel(shutdownCtx)
					newConf, newUpdated, err := c.Reload(newWatchCtx) // try to reload the configuration
					if err == nil {
						stopWatch() // stop watch the old config
						stopWatch = newStopWatch
						updated = newUpdated
						Reload(newConf)
						*c = *newConf
					} else {
						newStopWatch()
						logger.Error("Error reloading configuration. Configuration was left untouched.", "error", err)
					}
				}

			} else if sig == syscall.SIGTERM || sig == syscall.SIGINT {
				logger.Info("Termination signal received", "signal", sig)
				shutdown()
			} else {
				logger.Warn("Unknown signal received", "signal", sig)
			}

		case <-st.FatalErrorChan:
			logger.Warn("The store had a fatal error")
			shutdown()

		case <-st.KafkaErrorChan:
			logger.Warn("Store has received a Kafka error: resetting connection to Kafka")
			st.StopSendToKafka()
			st.SendToKafka(c.Kafka)

		case state := <-relpServer.StatusChan:
			switch state {
			case server.FinalStopped:
				logger.Info("The RELP service has been definitely halted")

			case server.Stopped:
				logger.Info("The RELP service has been stopped")
				relpServer.Conf = *c
				err := relpServer.Start()
				if err != nil {
					logger.Warn("The RELP service has failed to start", "error", err)
					relpServer.StopAndWait()
				}

			case server.Waiting:
				logger.Info("Waiting")
				go func() {
					time.Sleep(time.Duration(30) * time.Second)
					relpServer.EndWait()
				}()

			case server.Started:
				logger.Info("The RELP service has been started")
			}

		}

	}
}
