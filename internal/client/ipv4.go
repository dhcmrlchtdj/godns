package client

import (
	"context"
	"net"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

type Ipv4 struct {
	ip net.IP
}

func (ip *Ipv4) Resolve(ctx context.Context, question dns.Question, dnssec bool) ([]dns.RR, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "client.ipv4").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Logger()

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: question.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}
	rr.A = ip.ip

	logger.Debug().Msg("resolved")
	return []dns.RR{rr}, nil
}

func createIpv4Resolver(ctx context.Context, ip string) DnsResolver {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "client.ipv4").
		Logger()

	if client, found := resolverCache.Get(ip); found {
		logger.Trace().Msg("get resolver from cache")
		return client
	} else {
		client := &Ipv4{ip: net.ParseIP(ip)}
		resolverCache.Set(ip, client)
		logger.Trace().Msg("new resolver created")
		return client
	}
}
