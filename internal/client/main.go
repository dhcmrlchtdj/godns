package client

import (
	"context"

	"github.com/miekg/dns"
	"github.com/phuslu/shardmap"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/internal/config"
)

type DnsResolver interface {
	Resolve(ctx context.Context, question dns.Question, dnssec bool) ([]dns.RR, error)
}

var resolverCache = shardmap.New[string, DnsResolver](8) // nolint: gochecknoglobals

func GetByUpstream(ctx context.Context, upstream *config.Upstream) DnsResolver {
	if upstream == nil {
		return nil
	}

	if upstream.Block == "nodata" {
		return &BlockByNodata{}
	}
	if upstream.Block == "nxdomain" {
		return &BlockByNxdomain{}
	}
	if upstream.Ipv4 != "" {
		return createIpv4Resolver(ctx, upstream.Ipv4)
	}
	if upstream.Ipv6 != "" {
		return createIpv6Resolver(ctx, upstream.Ipv6)
	}
	if upstream.Udp != "" {
		return &Udp{server: upstream.Udp}
	}
	if upstream.Doh != "" {
		return createDohResolver(ctx, upstream.Doh, upstream.DohProxy)
	}

	zerolog.Ctx(ctx).Error().Str("module", "client.main").Msg("no upstream")

	return nil
}
