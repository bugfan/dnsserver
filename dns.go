package main

import (
	"strings"

	"github.com/bugfan/dnsserver/nameserver"
)

type DNSServer struct {
	*nameserver.Server
}

func NewDNSServer() *DNSServer {
	return &DNSServer{
		Server: nameserver.NewServer(),
	}
}

func (s *DNSServer) Start() {
	s.RunServer(":5354")
}

var testdomain = `
$ORIGIN example.com.
@                      3600 SOA   ns1.p30.dynect.net. (
                              zone-admin.dyndns.com.     ; address of responsible party
                              2016072701                 ; serial number
                              3600                       ; refresh period
                              600                        ; retry period
                              604800                     ; expire time
                              1800                     ) ; minimum ttl
                      86400 NS    ns1.p30.dynect.net.
                      86400 NS    ns2.p30.dynect.net.
                      86400 NS    ns3.p30.dynect.net.
                      86400 NS    ns4.p30.dynect.net.
                       3600 MX    10 mail.example.com.
                       3600 MX    20 vpn.example.com.
                       3600 MX    30 mail.example.com.
                         60 A     204.13.248.106
                       3600 TXT   "v=spf1 includespf.dynect.net ~all"
*                      3600 A     127.0.0.1
mail                  14400 A     204.13.248.106
vpn                      60 A     216.146.45.240
webapp                   60 A     216.146.46.10
webapp                   60 A     216.146.46.11
www                   43200 CNAME example.com.
rr                   43200 CNAME rr.example.com.
ww                   43200 CNAME w.w.example.com.
ww                   43200 CNAME w.w2.example.com.
ww                   43200 CNAME w.w3.example.com.
*.w                   43200 CNAME www.example.com.
*.w2                   43200 CNAME www.example.com.
*.w3                   43200 CNAME vpn.example.com.
`

func (s *DNSServer) LoadFromString(content, root string) {
	zone := nameserver.NewZone(root)
	zone.ParseReader(strings.NewReader(content))
	cache := nameserver.NewCacheProvider(zone)
	s.AddProvider(root, cache)
}

func main() {
	d := NewDNSServer()
	d.Start()
}
