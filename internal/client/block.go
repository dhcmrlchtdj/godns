package client

import (
	"context"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

type BlockByNodata struct{}

func (*BlockByNodata) Resolve(ctx context.Context, question dns.Question, dnssec bool) ([]dns.RR, error) {
	zerolog.Ctx(ctx).
		Debug().
		Str("module", "client.block.nodata").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Msg("resolved")
	return nil, nil
}

type BlockByNxdomain struct{}

func (*BlockByNxdomain) Resolve(ctx context.Context, question dns.Question, dnssec bool) ([]dns.RR, error) {
	zerolog.Ctx(ctx).
		Debug().
		Str("module", "client.block.nxdomain").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Msg("resolved")
	return nil, DnsResponseError{Rcode: dns.RcodeNameError}
}
