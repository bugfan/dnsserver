package nameserver

import "github.com/miekg/dns"

type DNSProvider interface {
	ServeDNS(w dns.ResponseWriter, req *dns.Msg) bool
}
type Handler struct {
	providers []DNSProvider
}

func NewHander() *Handler {
	return &Handler{}
}

func (h *Handler) AddProvider(p DNSProvider) {
	h.providers = append(h.providers, p)
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	for _, provider := range h.providers {
		if ok := provider.ServeDNS(w, req); ok {
			return
		}
	}
}
