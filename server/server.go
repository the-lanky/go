package lanky_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	llog "github.com/the-lanky/go/log"
	ltp "github.com/the-lanky/go/types"
)

// LankyServer represents a server that can be started and stopped.
type LankyServer interface {
	// Start starts the server.
	// It takes a context.Context and a channel to receive an os.Signal to gracefully shut down the server.
	Start(ctx context.Context, close chan os.Signal)
}

// Start starts the server and runs the API service.
// It listens for incoming requests and handles them accordingly.
// The server will run on the specified host and port.
// It also logs the server's address and provides instructions to stop the service.
// The function will gracefully shut down the server when a signal is received.
//
// Parameters:
//   - ctx: The context.Context object for managing the server's lifecycle.
//   - close: The channel to receive a signal for stopping the service.
func (s *ls) Start(ctx context.Context, close chan os.Signal) {
	apiFn := func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatalf("[❌] Failed start API Service: %+v", err)
		}
	}

	go apiFn()

	s.log.Infof("[🚀] API run on http://%s:%s", s.host, s.server.Addr)
	s.log.Info("[✨] Press CTRL+C to stop the service")
	s.gracefullShutdown(ctx, close)
}

// gracefullShutdown gracefully shuts down the server.
// It listens for the specified signals and waits for one of them to be received.
// Upon receiving a signal, it sets the server's keep-alive flag to false,
// creates a context with a timeout using the specified shutdown delay,
// and attempts to gracefully shut down the server using the Shutdown method.
// It then builds and logs a message indicating whether the shutdown was successful or not.
func (s *ls) gracefullShutdown(ctx context.Context, close chan os.Signal) {
	signal.Notify(
		close,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-close

	ctx, cancel := context.WithTimeout(ctx, s.conf.ShutdownDelay)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)

	err := s.server.Shutdown(ctx)
	s.buildMessage(
		err,
		"Successfully shutdown api service...",
		fmt.Sprintf("Failed to shutdown api service: %+v", err),
	)
}

type ls struct {
	server *http.Server
	conf   ltp.LankyServerConf
	host   string
	log    *logrus.Logger
}

// New creates a new instance of LankyServer with the given parameters.
// It initializes the server with the provided handler, configuration, and logger.
// If the logger is nil, it creates a new instance of llog with default settings.
// The server is configured with the provided host, address, and read timeout.
// If the configuration specifies a write timeout or idle timeout, they are also set on the server.
// The created LankyServer instance is returned.
func New(
	handler http.Handler,
	conf ltp.LankyServerConf,
	log *logrus.Logger,
) LankyServer {
	var (
		host = "localhost"
		addr = "8080"
		rto  = time.Second * 60
	)

	if log == nil {
		log = llog.NewInstance(
			llog.SetServiceName("API Service"),
			llog.SetIsProduction(false),
		)
	}

	if len(conf.Addr) > 0 {
		addr = conf.Addr
	}

	if len(conf.Host) > 0 {
		host = conf.Host
	}

	if conf.ReadTimeout > 0 {
		rto = conf.ReadTimeout
	}

	server := &http.Server{
		Addr:        fmt.Sprintf(":%s", addr),
		ReadTimeout: rto,
		Handler:     handler,
	}

	if conf.WriteTimeout > 0 {
		server.WriteTimeout = conf.WriteTimeout
	}

	if conf.IdleTimeout > 0 {
		server.IdleTimeout = conf.IdleTimeout
	}

	return &ls{
		host:   host,
		log:    log,
		conf:   conf,
		server: server,
	}
}

func (s *ls) buildMessage(err error, success, failed string) {
	if err == nil {
		s.log.Info(success)
	} else {
		s.log.Fatal(failed)
	}
}
