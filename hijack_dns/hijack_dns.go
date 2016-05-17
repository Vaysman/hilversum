package hijackdns

import (
	"github.com/spf13/viper"
	"github.com/miekg/dns"
	"time"
	"net"
	"math/rand"
	jww "github.com/spf13/jwalterweatherman"
	"os"
	"os/signal"
	"syscall"
)

func Run(viper *viper.Viper) {
	adr := "127.0.2.1:8053"
	net := "udp"

	pattern := dns.Fqdn(".")
	dns.HandleFunc(pattern, ServerHandler([]string{"8.8.8.8:53"}))
	go func() {
		err := dns.ListenAndServe(adr, net, nil)
		if err != nil {
			jww.FATAL.Fatalf("Failed to setup the " + adr + " server: %s\n", err.Error())
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	forever: for {
		select {
		case s := <-sig:
			jww.INFO.Printf("Signal (%s) received, stopping\n", s.String())
			break forever
		}
	}
}

// Returns an anonymous function configured to resolve DNS
// queries with a specific set of remote servers.
func ServerHandler(addresses []string) dns.HandlerFunc {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	// This is the actual handler
	return func(w dns.ResponseWriter, req *dns.Msg) {
		nameserver := addresses[randGen.Intn(len(addresses))]
		var protocol string

		switch t := w.RemoteAddr().(type) {
		default:
			jww.ERROR.Printf("Unsupported protocol %T\n", t)
			return
		case *net.UDPAddr:
			protocol = "udp"
		case *net.TCPAddr:
			protocol = "tcp"
		}

		for _, q := range req.Question {
			jww.INFO.Printf("Incoming request #%v: %s %s %v - using %s\n",
				req.Id,
				dns.ClassToString[q.Qclass],
				dns.TypeToString[q.Qtype],
				q.Name, nameserver)
		}

		for _, q := range req.Question {
			if q.Name == "ya.ru" {
				msg := new(dns.Msg)
				msg.SetReply(req)
				msg.Answer = make([]RR, 1)
				w.WriteMsg(msg)
				return
			}
		}
		c := new(dns.Client)
		c.Net = protocol
		resp, rtt, err := c.Exchange(req, nameserver)
		if err != nil {
			jww.ERROR.Printf("%s\n", err.Error())
			sendFailure(w, req)
			return
		}

		for _, a := range resp.Answer {
			jww.INFO.Printf("Incoming answer #%v: %v - from %s\n",
				resp.Id,
				a, nameserver)
		}
		jww.INFO.Printf("Request #%v: %.3d µs, server: %s(%s), size: %d bytes\n", resp.Id, rtt / 1e3, nameserver, c.Net, resp.Len())
		w.WriteMsg(resp)
	} // end of handler
}

func sendFailure(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetRcode(r, dns.RcodeServerFailure)
	w.WriteMsg(msg)
}