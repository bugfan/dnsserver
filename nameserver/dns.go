package nameserver

import (
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Mux      *dns.ServeMux
	handlers map[string]*Handler
	rTimeout time.Duration
	wTimeout time.Duration
}

func NewServer() *Server {
	s := &Server{
		handlers: make(map[string]*Handler),
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}
	s.Mux = dns.NewServeMux()
	return s
}

func (s *Server) AddProvider(pattern string, provider DNSProvider) {
	handler, ok := s.handlers[pattern]
	if !ok {
		handler = NewHander()
		s.handlers[pattern] = handler
		s.Mux.Handle(pattern, handler)
	}

	handler.AddProvider(provider)
}

func (s *Server) RunServer(listen string) {

	tcpServer := &dns.Server{
		Addr:         listen,
		Net:          "tcp",
		Handler:      s.Mux,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout,
	}

	udpServer := &dns.Server{
		Addr:         listen,
		Net:          "udp",
		Handler:      s.Mux,
		UDPSize:      65535,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout,
	}

	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		defer wait.Done()
		s.start(udpServer)
	}()
	go func() {
		defer wait.Done()
		s.start(tcpServer)
	}()

	wait.Wait()
}

func (s *Server) start(ds *dns.Server) {

	logrus.Infof("Start dns %s listener on %s", ds.Net, ds.Addr)
	err := ds.ListenAndServe()
	if err != nil {
		logrus.Errorf("Start dns %s listener on %s failed:%s", ds.Net, ds.Addr, err.Error())
	}

}

func ClientIP(w dns.ResponseWriter) net.IP {
	switch inst := w.RemoteAddr().(type) {
	case *net.TCPAddr:
		return inst.IP
	case *net.UDPAddr:
		return inst.IP
	}
	return nil
}
