package nameserver

import (
	"time"

	"github.com/miekg/dns"
	cache "github.com/patrickmn/go-cache"
)

type Resolver interface {
	Resolve(dns.Question) []dns.RR
}

type CachedProvider struct {
	cache    *cache.Cache
	Resolver Resolver
}

func NewCacheProvider(r Resolver) DNSProvider {
	return &CachedProvider{
		cache:    cache.New(2*time.Hour, 10*time.Minute),
		Resolver: r,
	}
}

func (p *CachedProvider) ServeDNS(w dns.ResponseWriter, req *dns.Msg) bool {
	m := new(dns.Msg)
	m.SetReply(req)
	ok := false
	for _, q := range req.Question {
		res := p.Resolve(q)
		if len(res) <= 0 {
			continue
		}

		for _, r := range res {
			m.Answer = append(m.Answer, r)
			ok = true
		}

	}
	w.WriteMsg(m)
	return ok
}

func (p *CachedProvider) Resolve(q dns.Question) []dns.RR {
	if res, ok := p.cache.Get(q.String()); ok {
		return res.([]dns.RR)
	}

	r := p.Resolver.Resolve(q)
	if r != nil && len(r) != 0 {
		p.cache.Set(q.String(), r, time.Duration(r[0].Header().Ttl)*time.Second)
	}
	return r
}
