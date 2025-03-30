package client

import "github.com/miekg/dns"

type DnsResponseError struct {
	Rcode int
}

func (e DnsResponseError) Error() string {
	return dns.RcodeToString[e.Rcode]
}
