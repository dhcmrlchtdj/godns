package server

import (
	"context"
	"errors"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/internal/client"
	"github.com/dhcmrlchtdj/godns/internal/util"
)

func (s *DnsServer) handleRequest(w dns.ResponseWriter, request *dns.Msg) {
	loggerWithId := zerolog.Ctx(s.ctx).
		With().
		Uint16("request_id", request.Id).
		Logger()
	ctx := loggerWithId.WithContext(s.ctx)

	logger := loggerWithId.
		With().
		Str("module", "server.handler").
		Logger()

	start := time.Now()

	reply := new(dns.Msg)
	reply.SetReply(request)
	if edns := request.IsEdns0(); edns != nil {
		reply.SetEdns0(4096, true)
	}

	logger.Trace().
		Str("opcode", dns.OpcodeToString[request.Opcode]).
		Msg("receive request")
	if request.Opcode == dns.OpcodeQuery {
		s.query(ctx, reply)
	} else {
		reply.Rcode = dns.RcodeNotImplemented
	}

	err := w.WriteMsg(reply)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to write reply")
	}

	latency := time.Since(start)
	logger.Trace().Dur("latency", latency).Send()
}

func (s *DnsServer) query(ctx context.Context, reply *dns.Msg) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.handler").
		Logger()

	if len(reply.Question) != 1 {
		logger.Debug().
			Int("question", len(reply.Question)).
			Msg("format error")
		reply.Rcode = dns.RcodeFormatError
		return
	}

	question := reply.Question[0]
	logger.Info().
		Str("name", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Bool("dnssec", reply.IsEdns0() != nil).
		Msg("query")

	// from cache
	cacheKey := question.String()
	answer, err := s.cacheGet(ctx, cacheKey)
	if err != nil {
		var errRcode client.ErrDnsResponse
		if errors.As(err, &errRcode) {
			reply.Rcode = errRcode.Rcode
		} else {
			reply.Rcode = dns.RcodeServerFailure
		}
		logger.Trace().Msg("from cache")
		return
	} else if answer != nil {
		reply.Answer = answer
		logger.Trace().Msg("from cache")
		return
	}

	deferred := util.MakeDeferred[cachedAnswer]()
	s.cacheSet(ctx, cacheKey, deferred)

	upstream := s.router.search(ctx, question.Name, question.Qtype)

	// no upstream
	if upstream == nil {
		logger.Trace().Msg("no upstream")
		reply.Rcode = dns.RcodeNotImplemented
		s.cacheReject(ctx, cacheKey, reply.Rcode)
		return
	}
	// no resolver
	resolver := client.GetByUpstream(ctx, upstream)
	if resolver == nil {
		logger.Error().Msg("no resolver")
		reply.Rcode = dns.RcodeNotImplemented
		s.cacheReject(ctx, cacheKey, reply.Rcode)
		return
	}

	// from upstream
	ans, err := resolver.Resolve(ctx, question, reply.IsEdns0() != nil)
	if err == nil {
		reply.Answer = ans
		logger.Trace().Msg("resolved")
		s.cacheResolve(ctx, cacheKey, ans)
	} else {
		var errRcode client.ErrDnsResponse
		if errors.As(err, &errRcode) {
			reply.Rcode = errRcode.Rcode
			logger.Debug().
				Str("Rcode", dns.RcodeToString[reply.Rcode]).
				Msg("resolved")
		} else {
			reply.Rcode = dns.RcodeServerFailure
			logger.Error().Stack().Err(err).Msg("unknown error")
		}
		s.cacheReject(ctx, cacheKey, reply.Rcode)
	}
}
