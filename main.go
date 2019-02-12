package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/jonathanbeber/dns-proxy/config"
	"github.com/miekg/dns"
)

type handler struct {
	client dns.Client
	config config.Config
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	r, _, err := h.client.Exchange(r, h.config.UpstreamServer+h.config.UpstreamPort)
	if err != nil {
		log.Printf("failed to communicate with upstream: %s", err)
		return
	}
	w.WriteMsg(r)
}

func main() {
	cfg := config.Config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Panic("Failed to parse config")
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	signal.Notify(signalChan, syscall.SIGINT)

	c := new(dns.Client)
	c.Net = "tcp-tls"
	c.Dialer = &net.Dialer{
		Timeout: 200 * time.Millisecond,
	}

	server_udp := &dns.Server{Addr: ":53", Net: "tcp"}
	server_tcp := &dns.Server{Addr: ":53", Net: "udp"}
	go server_udp.ListenAndServe()
	go server_tcp.ListenAndServe()
	dns.HandleFunc(".", handleFunc)

	sig := <-signalChan
	log.Printf("Received signal: %q, shutting down..", sig.String())
	shutdownServer(server_tcp)
	shutdownServer(server_udp)
}

func shutdownServer(s *dns.Server) {
	if err := s.Shutdown(); err != nil {
		log.Printf("Failed to shutdown server %s", s.Net)
	}
}
