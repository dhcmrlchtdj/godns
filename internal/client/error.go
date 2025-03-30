package client

import "github.com/miekg/dns"

type ErrDnsResponse struct {
	Rcode int
}

func (e ErrDnsResponse) Error() string {
	return dns.RcodeToString[e.Rcode]
}
