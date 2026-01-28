package ipin_test

import (
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
	ipin "github.com/wenerme/coredns-ipin"
	"golang.org/x/net/context"
)

func TestWhoami(t *testing.T) {
	wh := ipin.IpInName{Ttl: uint32(86400)}

	tests := []struct {
		qname        string
		qtype        uint16
		expectedCode int
		expectedName string // expected record name
		expectedErr  error
		expectedIP   string // for A records
	}{
		{
			qname:        "192-168-1-2.example.org",
			qtype:        dns.TypeA,
			expectedCode: dns.RcodeSuccess,
			expectedName: "192-168-1-2.example.org.",
			expectedErr:  nil,
			expectedIP:   "192.168.1.2",
		},
		{
			qname:        "prefix-192-168-1-2.example.org",
			qtype:        dns.TypeA,
			expectedCode: dns.RcodeSuccess,
			expectedName: "prefix-192-168-1-2.example.org.",
			expectedErr:  nil,
			expectedIP:   "192.168.1.2",
		},
	}

	ctx := context.TODO()

	for i, tc := range tests {
		req := new(dns.Msg)
		req.SetQuestion(dns.Fqdn(tc.qname), tc.qtype)

		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		code, err := wh.ServeDNS(ctx, rec, req)

		if err != tc.expectedErr {
			t.Errorf("Test %d: Expected error %v, but got %v", i, tc.expectedErr, err)
		}
		if code != int(tc.expectedCode) {
			t.Errorf("Test %d: Expected status code %d, but got %d", i, tc.expectedCode, code)
		}

		if len(rec.Msg.Answer) == 0 {
			t.Errorf("Test %d: Expected answer record but got none", i)
			continue
		}

		// Check record name
		actual := rec.Msg.Answer[0].Header().Name
		if actual != tc.expectedName {
			t.Errorf("Test %d: Expected record name %s, but got %s", i, tc.expectedName, actual)
		}

		// Check A record IP address
		aRecord, ok := rec.Msg.Answer[0].(*dns.A)
		if !ok {
			t.Errorf("Test %d: Expected A record, but got %T", i, rec.Msg.Answer[0])
		} else {
			actualIP := aRecord.A.String()
			if actualIP != tc.expectedIP {
				t.Errorf("Test %d: Expected IP %s, but got %s", i, tc.expectedIP, actualIP)
			}
		}
	}
}
