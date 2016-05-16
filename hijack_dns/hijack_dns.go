package hijackdns

import (
	"github.com/spf13/viper"
	"github.com/miekg/dns"
	"time"
	"net"
	"math/rand"
	jww "github.com/spf13/jwalterweatherman"
)

type handler func(dns.ResponseWriter, *dns.Msg)

func Run(viper *viper.Viper) {
	net := "127.0.0.1:8053"
	err := dns.ListenAndServe("udp", net, nil)
	if err != nil {
		jww.FATAL.Fatalf("Failed to setup the " + net + " server: %s\n", err.Error())
	}
}

// Returns an anonymous function configured to resolve DNS
// queries with a specific set of remote servers.
func ServerHandler(addresses []string) handler {
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

		c := new(dns.Client)
		c.Net = protocol
		resp, rtt, err := c.Exchange(req, nameserver)

		Redo:
		switch {
		case err != nil:
			jww.ERROR.Printf("%s\n", err.Error())
			sendFailure(w, req)
			return
		case req.Id != resp.Id:
			jww.ERROR.Printf("Id mismatch: %v != %v\n", req.Id, resp.Id)
			sendFailure(w, req)
			return
		case resp.MsgHdr.Truncated && protocol != "tcp":
			jww.WARN.Printf("Truncated answer for request %v, retrying TCP\n", req.Id)
			c.Net = "tcp"
			resp, rtt, err = c.Exchange(req, nameserver)
			goto Redo
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