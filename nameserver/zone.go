package nameserver

import (
	"io"
	"strings"

	"github.com/miekg/dns"
)

type Zone struct {
	in        map[string]map[uint16][]dns.RR
	origin    string
	enableAny bool
}

func NewZone(origin string) *Zone {
	return &Zone{
		in:     make(map[string]map[uint16][]dns.RR),
		origin: origin,
	}
}

func (z *Zone) ParseReader(r io.Reader) {
	parser := dns.NewZoneParser(r, z.origin, "")
	for rr, next := parser.Next(); next; rr, next = parser.Next() {
		z.AddRR(rr)
	}
}

func (z *Zone) ParseString(s string) {
	z.ParseReader(strings.NewReader(s))
}

func (z *Zone) AddRR(rr dns.RR) {
	h := rr.Header()
	if z.in[h.Name] == nil {
		z.in[h.Name] = make(map[uint16][]dns.RR)
	}
	z.in[h.Name][h.Rrtype] = append(z.in[h.Name][h.Rrtype], rr)
}

func (z *Zone) Resolve(q dns.Question) []dns.RR {
	if q.Qclass != dns.ClassINET {
		return nil
	}

	if !strings.HasSuffix(q.Name, z.origin) {
		return nil
	}
	resolved := make(map[string][]dns.RR)
	return z.inResolve(q.Name, q.Qtype, resolved)
}
func (z *Zone) inResolve(name string, qtype uint16, resolved map[string][]dns.RR) []dns.RR {
	var rrs []dns.RR
	res := z.inResolveNormal(name, qtype)
	if res != nil {
		resolved[name] = res
		return res
	}
	res = z.inResolveNormal(name, dns.TypeCNAME)
	if res != nil {
		resolved[name] = res
		rrs = res
		for _, rr := range res {
			cname, ok := rr.(*dns.CNAME)
			if !ok {
				continue
			}
			if _, ok = resolved[cname.Target]; !ok {
				subrr := z.inResolve(cname.Target, qtype, resolved)
				if len(subrr) > 0 {
					rrs = append(rrs, subrr...)
				}
			}
		}
		return rrs
	}

	res = z.inResolveWildcard(name, qtype)
	if res != nil {
		resolved[name] = res
		return res
	}
	res = z.inResolveWildcard(name, dns.TypeCNAME)
	if res != nil {
		resolved[name] = res
		rrs = res
		for _, rr := range res {
			cname, ok := rr.(*dns.CNAME)
			if !ok {
				continue
			}
			if _, ok = resolved[cname.Target]; !ok {
				subrr := z.inResolve(cname.Target, qtype, resolved)
				if len(subrr) > 0 {
					rrs = append(rrs, subrr...)
				}
			}
		}
		return rrs
	}
	return nil
}

func (z *Zone) inResolveNormal(name string, qtype uint16) []dns.RR {
	if z.in[name] != nil {
		if qtype == dns.TypeANY {
			rrs := make([]dns.RR, 0, 10)
			for _, v := range z.in[name] {
				rrs = append(rrs, v...)
			}
			return rrs
		}
		if z.in[name][qtype] != nil && len(z.in[name][qtype]) > 0 {
			return z.in[name][qtype]
		}
	}
	return nil
}
func (z *Zone) inResolveWildcard(name string, qtype uint16) []dns.RR {

	for _, w := range wildcardNames(name, z.origin) {
		if z.in[w][qtype] != nil && len(z.in[w][qtype]) > 0 {
			rrs := make([]dns.RR, 0, len(z.in[w][qtype]))
			for _, rr := range z.in[w][qtype] {
				newrr := dns.Copy(rr)
				newrr.Header().Name = name
				rrs = append(rrs, newrr)
			}
			return rrs
		}
	}
	return nil
}

func wildcardNames(name, origin string) []string {
	name = strings.TrimSuffix(name, origin)
	ds := strings.Split(name, ".")
	wild := make([]string, 0, len(ds))
	res := make([]string, 0, len(ds))
	for i := 0; i < len(ds); i++ {
		if ds[i] == "" {
			continue
		}
		wild = append(wild, "*")
		names := append(wild, ds[i+1:]...)
		wildName := strings.Join(names, ".") + origin
		res = append(res, wildName)
	}
	return res
}

type CacheZone struct {
	*Zone
	cache DNSProvider
}

func NewCacheZone(origin string) *CacheZone {
	zone := NewZone(origin)
	cache := NewCacheProvider(zone)
	return &CacheZone{
		Zone:  zone,
		cache: cache,
	}
}
func (p *CacheZone) ServeDNS(w dns.ResponseWriter, req *dns.Msg) bool {
	return p.cache.ServeDNS(w, req)
}

func (p *CacheZone) Resolve(q dns.Question) []dns.RR {
	// return p.cache.Resolve(q)
	return nil
}
