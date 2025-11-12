package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "net/http/pprof" // #nosec

	"github.com/miekg/dns"
	"github.com/phuslu/shardmap"
	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/internal/config"
)

type DnsServer struct {
	dnsServer     *dns.Server
	pprofServer   *http.Server
	pprofListener net.Listener
	ctx           context.Context // nolint: containedctx
	router        *router
	cache         *shardmap.Map[string, *deferredAnswer]
	Config        config.Config
}

func NewDnsServer(ctx context.Context) *DnsServer {
	server := new(DnsServer)
	server.cache = shardmap.New[string, *deferredAnswer](64)

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	logger = logger.Hook(LogIdHook{})
	server.ctx = logger.WithContext(ctx)

	go func() {
		<-ctx.Done()
		logger.Info().
			Str("module", "server.main").
			Msg("DNS server is stopping")

		server.shutdownDNS()
		server.shutdownPprof()
	}()

	return server
}

func (s *DnsServer) SetupRouter() {
	zerolog.Ctx(s.ctx).
		Debug().
		Str("module", "server.main").
		Msg("loading config")
	s.router = newRouter()
	s.router.addRules(s.ctx, s.Config.Rule, false)
}

func (s *DnsServer) SetupServer() {
	dnsMux := dns.NewServeMux()
	s.dnsServer = &dns.Server{
		Addr:    net.JoinHostPort(s.Config.Host, strconv.Itoa(s.Config.Port)),
		Net:     "udp",
		Handler: dnsMux,
		NotifyStartedFunc: func() {
			addr := s.dnsServer.PacketConn.LocalAddr()
			zerolog.Ctx(s.ctx).
				Info().
				Str("module", "server.main").
				Str("log_level", s.Config.LogLevel).
				Str("server_addr", addr.String()).
				Msg("DNS server is running")

			s.router.addRules(s.ctx, s.Config.Rule, true)
		},
	}
	dnsMux.HandleFunc(".", s.handleRequest)
}

func (s *DnsServer) SetupPprof() {
	pprofMux := http.DefaultServeMux
	http.DefaultServeMux = http.NewServeMux()
	s.pprofServer = &http.Server{
		Handler:           pprofMux,
		ReadHeaderTimeout: 10 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return s.ctx
		},
	}
	netListener, err := net.Listen("tcp", s.Config.Host+":0") // nolint: noctx
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
	s.pprofListener = netListener
}

func (s *DnsServer) Start() {
	var wg sync.WaitGroup

	wg.Go(s.cleanupExpiredCache)
	wg.Go(s.startDNS)
	wg.Go(s.startPprof)

	wg.Wait()
}

func (s *DnsServer) startDNS() {
	err := s.dnsServer.ListenAndServe()
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
}

func (s *DnsServer) shutdownDNS() {
	err := s.dnsServer.Shutdown()
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
}

func (s *DnsServer) startPprof() {
	if zerolog.GlobalLevel() != zerolog.TraceLevel {
		return
	}

	zerolog.Ctx(s.ctx).
		Info().
		Str("module", "server.main").
		Str("pprof_addr", "http://"+s.pprofListener.Addr().String()+"/debug/pprof/").
		Msg("pprof is running")

	err := s.pprofServer.Serve(s.pprofListener)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			zerolog.Ctx(s.ctx).
				Error().
				Str("module", "server.main").
				Stack().
				Err(err).
				Send()
			panic(err)
		}
	}
}

// nolint: contextcheck
func (s *DnsServer) shutdownPprof() {
	err := s.pprofServer.Shutdown(context.Background())
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
}

///

type LogIdHook struct{}

func (LogIdHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	e.Str("logId", xid.New().String())
}
