package hijackdns

import (
	"github.com/spf13/viper"
	"github.com/miekg/dns"
	"net"
	jww "github.com/spf13/jwalterweatherman"
)

const ttl = 1

func Run(viper *viper.Viper) {
	adr := "127.0.2.1:8053"
	net := "udp"

	Configure(viper)
	go func() {
		err := dns.ListenAndServe(adr, net, nil)
		if err != nil {
			jww.FATAL.Fatalf("Failed to setup the " + adr + " server: %s\n", err.Error())
		}
	}()
}

func Configure(viper *viper.Viper) {
	addrs := map[string]string{
		"y.ru.": "1.1.1.1",
		"i.ru.": "2.2.2.2",
		"ya.ru.": "3.3.3.3",
		"i.co.": "4.4.4.4",
	}
	nameserver := "8.8.8.8:53"
	pattern := dns.Fqdn(".")
	dns.HandleRemove(pattern)
	dns.HandleFunc(pattern, serverHandler(addrs, nameserver))
	jww.DEBUG.Print("Configuring DNS Hijack")
}

// Returns an anonymous function configured to resolve DNS
// queries with a specific set of remote servers.
func serverHandler(addresses map[string]string, nameserver string) dns.HandlerFunc {
	// This is the actual handler
	return func(w dns.ResponseWriter, req *dns.Msg) {
		for _, q := range req.Question {
			jww.INFO.Printf("Incoming request #%v: %s %s %v - using %s",
				req.Id,
				dns.ClassToString[q.Qclass],
				dns.TypeToString[q.Qtype],
				q.Name, nameserver)
		}

		for _, q := range req.Question {
			if addr, ok := addresses[q.Name]; ok {
				msg := new(dns.Msg)
				msg.SetReply(req)
				a := &dns.A{
					Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl},
					A: net.ParseIP(addr).To4(),
				}
				msg.Answer = append(msg.Answer, a)
				w.WriteMsg(msg)
				return
			}
		}
		c := new(dns.Client)
		resp, rtt, err := c.Exchange(req, nameserver)
		if err != nil {
			jww.ERROR.Printf("%s", err.Error())
			sendFailure(w, req)
			return
		}

		jww.INFO.Printf("Request #%v: %.3d µs, server: %s(%s), size: %d bytes", resp.Id, rtt / 1e3, nameserver, c.Net, resp.Len())
		w.WriteMsg(resp)
	} // end of handler
}

func sendFailure(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetRcode(r, dns.RcodeServerFailure)
	w.WriteMsg(msg)
}