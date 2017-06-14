package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var (
	_resolve string
	_listen  string
)

func init() {
	flag.StringVar(&_resolve, "r", "", "dns server")
	flag.StringVar(&_listen, "l", "127.0.0.1:53", "listen address")
	flag.Parse()
}

func filter(block []string, qs []dns.Question) []dns.Question {
	as := []dns.Question{}
next:
	for _, q := range qs {
		for _, b := range block {
			if strings.HasSuffix(q.Name, b) {
				continue next
			}
		}
		as = append(as, q)
	}
	return as
}

func main() {
	if _resolve == "" {
		flag.Usage()
		return
	}
	block := []string{}
	{
		fd, err := os.Open("dnsb.block")
		if err != nil {
			log.Fatal(err)
		}
		bin := bufio.NewReader(fd)
		for {
			line, _ := bin.ReadString('\n')
			if line == "" {
				break
			}
			block = append(block, strings.TrimRight(line, "\n")+".")
		}
	}
	var c dns.Client
	s := dns.Server{
		Net:  "udp",
		Addr: _listen,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			if r.Question = filter(block, r.Question); len(r.Question) == 0 {
				w.WriteMsg(r)
				return
			}
			in, _, err := c.Exchange(r, _resolve)
			if err != nil {
				dns.HandleFailed(w, r)
				return
			}
			w.WriteMsg(in)
		}),
	}
	log.Fatal(s.ListenAndServe())
}
