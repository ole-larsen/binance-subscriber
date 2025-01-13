package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ole-larsen/binance-subscriber/internal/httpserver"
	"github.com/ole-larsen/binance-subscriber/internal/httpserver/router"
	"github.com/ole-larsen/binance-subscriber/internal/log"
	"github.com/ole-larsen/binance-subscriber/internal/poller"
	"github.com/ole-larsen/binance-subscriber/internal/server/config"
	"github.com/ole-larsen/binance-subscriber/internal/storage"
)

var (
	logger = log.NewLogger("info", log.DefaultBuildLogger)
)

// Server represents the server instance, encapsulating settings,
// logger, signal handling, and storage and gRPC server components.
type Server struct {
	http     *httpserver.HTTPServer
	poller   *poller.BinancePoller
	storage  storage.Storage
	settings *config.Config
	logger   *log.Logger
	signal   chan os.Signal
	done     chan struct{}
}

// NewServer creates and returns a new Server instance with default logger settings.
func NewServer() *Server {
	return &Server{
		logger: logger,
	}
}

var SetupFunc = Setup

// Setup initializes the server with provided configuration settings and sets up
// storage. Returns an error if initialization fails.
func Setup(settings *config.Config) (*Server, error) {
	s := NewServer()

	if err := s.Init(storage.NewMemStorage(), settings, make(chan os.Signal, 1), make(chan struct{})); err != nil {
		return nil, err
	}

	return s, nil
}

// Run starts the server and begins listening for shutdown signals. It runs the gRPC server
// and handles shutdown on receiving system interrupt signals like SIGINT or SIGTERM.
func (s *Server) Run(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	defer close(s.signal)

	// shutdown workers
	go func() {
		signal.Notify(s.signal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	}()

	go func(signal chan os.Signal, done chan struct{}) {
		<-signal
		close(done)
		s.logger.Infow("...graceful server shutdown")
	}(s.signal, s.done)

	port := s.settings.Port
	host := s.settings.Host

	if err := s.poller.Connect(ctx); err != nil {
		s.logger.Errorln(err)
		return
	}

	defer func() {
		if err := s.poller.Close(); err != nil {
			s.logger.Errorln(err)
			return
		}
	}()

	if err := s.poller.Subscribe(s.settings.Instruments); err != nil {
		s.logger.Errorln(err)
		return
	}

	go func() {
		if err := s.poller.Read(); err != nil {
			s.logger.Errorln(err)
			return
		}
	}()

	s.logger.Infow("...starting server",
		"host", host,
		"port", port,
		"goroutines", runtime.NumGoroutine(),
	)

	go func() {
		if err := s.http.ListenAndServe(); err != nil {
			s.logger.Errorln(err)
		}
	}()

	for {
		select {
		case message, ok := <-s.poller.GetMsg():
			if ok {
				var resp poller.DepthMessage

				err := json.Unmarshal(message, &resp)
				if err != nil {
					fmt.Println("Error unmarshalling JSON:", err)
					return
				}

				if len(resp.Data.Asks) > 0 && len(resp.Data.Bids) > 0 {
					s.storage.Set(storage.Data{
						Symbol: resp.Data.Symbol,
						Bid:    resp.Data.Bids[0][0],
						Ask:    resp.Data.Asks[0][0],
					})
				}
			}
		case <-s.done:
			if err := s.poller.Unsubscribe(s.settings.Instruments); err != nil {
				s.logger.Errorln(err)
				return
			}

			s.logger.Infow("...stop server",
				"goroutines", runtime.NumGoroutine(),
			)

			return
		case <-ctx.Done():
			s.logger.Infow("stop server by ctx")
			return
		}
	}
}

// Init initializes the server with the given settings, signal channels.
// Returns an error if any component is missing.
func (s *Server) Init(
	store storage.Storage,
	settings *config.Config,
	sgnl chan os.Signal,
	done chan struct{},
) error {
	s.SetSettings(settings).
		SetSignal(sgnl).
		SetDone(done).
		SetStorage(store)

	if s.settings == nil {
		return NewError(errors.New("config is missing"))
	}

	if s.storage == nil {
		return NewError(errors.New("storage is missing"))
	}

	if s.signal == nil {
		return NewError(errors.New("signal is missing"))
	}

	if s.done == nil {
		return NewError(errors.New("done is missing"))
	}

	r := router.NewMux().
		SetStorage(store).
		SetMiddlewares().
		SetHandlers()

	s.
		SetHTTPServer(httpserver.NewHTTPServer().
			SetHost(s.settings.Host).
			SetPort(s.settings.Port).
			SetRouter(r))

	s.SetBinancePoller(poller.NewBinancePoller())

	if s.http == nil {
		return NewError(errors.New("http server is missing"))
	}

	if s.http != nil {
		if s.http.GetPort() == 0 {
			return NewError(errors.New("http server port is missing"))
		}

		if s.http.GetRouter() == nil {
			return NewError(errors.New("http server router is missing"))
		}
	}

	if s.poller == nil {
		return NewError(errors.New("ws client is missing"))
	}

	return nil
}

// SetSettings sets the server configuration.
func (s *Server) SetSettings(settings *config.Config) *Server {
	s.settings = settings
	return s
}

// SetSignal sets the signal channel for handling OS signals.
func (s *Server) SetSignal(sgnl chan os.Signal) *Server {
	s.signal = sgnl
	return s
}

// SetDone sets the done channel to signal when the server should stop.
func (s *Server) SetDone(done chan struct{}) *Server {
	s.done = done
	return s
}

func (s *Server) SetHTTPServer(hs *httpserver.HTTPServer) *Server {
	s.http = hs
	return s
}

func (s *Server) SetBinancePoller(ws *poller.BinancePoller) *Server {
	s.poller = ws
	return s
}

func (s *Server) SetStorage(store storage.Storage) *Server {
	s.storage = store
	return s
}

// GetSettings retrieves the server configuration settings.
func (s *Server) GetSettings() *config.Config {
	return s.settings
}

// GetSignal retrieves the signal channel used by the server.
func (s *Server) GetSignal() chan os.Signal {
	return s.signal
}

// GetDone retrieves the done channel used by the server.
func (s *Server) GetDone() chan struct{} {
	return s.done
}

// GetLogger retrieves the logger used by the server.
func (s *Server) GetLogger() *log.Logger {
	return s.logger
}

// GetStorage retrives the storage.
func (s *Server) GetStorage() *storage.MemStorage {
	if s.storage == nil {
		return nil
	}

	store, ok := s.storage.(*storage.MemStorage)
	if !ok {
		return nil
	}

	return store
}

func (s *Server) GetHTTPServer() *httpserver.HTTPServer {
	return s.http
}

func (s *Server) GetBinancePoller() *poller.BinancePoller {
	return s.poller
}
