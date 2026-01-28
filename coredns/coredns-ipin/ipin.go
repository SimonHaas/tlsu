// Package whoami implements a plugin that returns details about the resolving
// querying it.
package ipin

import (
	"net"

	"github.com/coredns/coredns/request"

	"regexp"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

const Name = "ipin"

type IpInName struct {
	// When process failed, will call next plugin
	Fallback bool
	Ttl      uint32
	Next     plugin.Handler
}

var regIpDash = regexp.MustCompile(`^(?:[a-z]+-)?(\d{1,3}-\d{1,3}-\d{1,3}-\d{1,3})\.`)

func (self IpInName) Name() string { return Name }
func (self IpInName) Resolve(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (*dns.Msg, int, error) {
	state := request.Request{W: w, Req: r}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Compress = true
	a.Authoritative = true

	matches := regIpDash.FindStringSubmatch(state.QName())
	if len(matches) > 1 {
		ip := matches[1]
		ip = strings.Replace(ip, "-", ".", -1)

		var rr dns.RR
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(), Ttl: self.Ttl}
		rr.(*dns.A).A = net.ParseIP(ip).To4()

		a.Answer = []dns.RR{rr}
	}

	return a, 0, nil
}
func (self IpInName) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	a, i, err := self.Resolve(ctx, w, r)
	if err != nil {
		return i, err
	}

	if self.Fallback && len(a.Answer) == 0 {
		return self.Next.ServeDNS(ctx, w, r)
	} else {
		return 0, w.WriteMsg(a)
	}
}
