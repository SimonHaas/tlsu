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
		qname         string
		qtype         uint16
		expectedCode  int
		expectedReply []string // ownernames for the records in the additional section.
		expectedTtl   uint32
		expectedErr   error
		expectedIP    string // for A records
		expectedIPv6  string // for AAAA records
		expectedPort  uint16 // for SRV records
	}{
		{
			qname:         "192-168-1-2-80.example.org",
			qtype:         dns.TypeA,
			expectedCode:  dns.RcodeSuccess,
			expectedReply: []string{"192-168-1-2-80.example.org.", "_port.192-168-1-2-80.example.org."},
			expectedTtl:   uint32(86400),
			expectedErr:   nil,
			expectedIP:    "192.168.1.2",
			expectedPort:  80,
		},
		{
			qname:         "prefix-192-168-1-2.example.org",
			qtype:         dns.TypeA,
			expectedCode:  dns.RcodeSuccess,
			expectedReply: []string{"prefix-192-168-1-2.example.org."},
			expectedTtl:   uint32(86400),
			expectedErr:   nil,
			expectedIP:    "192.168.1.2",
		},
		{
			qname:         "prefix-192-168-1-2-80.example.org",
			qtype:         dns.TypeA,
			expectedCode:  dns.RcodeSuccess,
			expectedReply: []string{"prefix-192-168-1-2-80.example.org.", "_port.prefix-192-168-1-2-80.example.org."},
			expectedTtl:   uint32(86400),
			expectedErr:   nil,
			expectedIP:    "192.168.1.2",
			expectedPort:  80,
		},
		{
			qname:         "fe80--1ff-fe23-4567-890a.example.org",
			qtype:         dns.TypeAAAA,
			expectedCode:  dns.RcodeSuccess,
			expectedReply: []string{"fe80--1ff-fe23-4567-890a.example.org."},
			expectedTtl:   uint32(86400),
			expectedErr:   nil,
			expectedIPv6:  "fe80::1ff:fe23:4567:890a",
		},
		{
			qname:         "prefix-fe80--1ff-fe23-4567-890a.example.org",
			qtype:         dns.TypeAAAA,
			expectedCode:  dns.RcodeSuccess,
			expectedReply: []string{"prefix-fe80--1ff-fe23-4567-890a.example.org."},
			expectedTtl:   uint32(86400),
			expectedErr:   nil,
			expectedIPv6:  "fe80::1ff:fe23:4567:890a",
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
		if len(tc.expectedReply) != 0 {
			actual := rec.Msg.Answer[0].Header().Name
			expected := tc.expectedReply[0]
			if actual != expected {
				t.Errorf("Test %d: Expected answer %s, but got %s", i, expected, actual)
			}

			// Check A record IP address
			if tc.qtype == dns.TypeA && tc.expectedIP != "" {
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

			// Check AAAA record IP address
			if tc.qtype == dns.TypeAAAA && tc.expectedIPv6 != "" {
				aaaaRecord, ok := rec.Msg.Answer[0].(*dns.AAAA)
				if !ok {
					t.Errorf("Test %d: Expected AAAA record, but got %T", i, rec.Msg.Answer[0])
				} else {
					actualIP := aaaaRecord.AAAA.String()
					if actualIP != tc.expectedIPv6 {
						t.Errorf("Test %d: Expected IPv6 %s, but got %s", i, tc.expectedIPv6, actualIP)
					}
				}
			}

			if len(tc.expectedReply) > 1 {
				if len(rec.Msg.Extra) > 0 {
					actual = rec.Msg.Extra[0].Header().Name
					expected = tc.expectedReply[1]
					if actual != expected {
						t.Errorf("Test %d: Expected answer %s, but got %s", i, expected, actual)
					}

					if rec.Msg.Extra[0].Header().Ttl != tc.expectedTtl {
						t.Errorf("Test %d: Expected answer %d, but got %d", i, tc.expectedTtl, rec.Msg.Extra[0].Header().Ttl)
					}

					// Check SRV record port
					if tc.expectedPort != 0 {
						srvRecord, ok := rec.Msg.Extra[0].(*dns.SRV)
						if !ok {
							t.Errorf("Test %d: Expected SRV record, but got %T", i, rec.Msg.Extra[0])
						} else {
							if srvRecord.Port != tc.expectedPort {
								t.Errorf("Test %d: Expected port %d, but got %d", i, tc.expectedPort, srvRecord.Port)
							}
						}
					}
				} else {
					t.Errorf("Test %d: Expected answer %s, but got no response", i, tc.expectedReply[1])
				}
			}
		}
	}
}
