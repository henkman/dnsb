package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

const (
	BLOCK_FILE = "dnsb.block"
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

func filter(block []string, qs *[]dns.Question) {
	i := 0
next:
	for i < len((*qs)) {
		for _, b := range block {
			if strings.HasSuffix((*qs)[i].Name, b) {
				(*qs)[i] = (*qs)[len((*qs))-1]
				(*qs) = (*qs)[:len((*qs))-1]
				continue next
			}
		}
		i++
	}
}

func read_blocks() ([]string, error) {
	block := []string{}
	fd, err := os.Open(BLOCK_FILE)
	if err != nil {
		return nil, err
	}
	bin := bufio.NewReader(fd)
	for {
		line, _ := bin.ReadString('\n')
		if line == "" {
			break
		}
		block = append(block, strings.TrimRight(line, "\n")+".")
	}
	return block, nil
}

func main() {
	if _resolve == "" {
		flag.Usage()
		return
	}
	block, err := read_blocks()
	if err != nil {
		log.Fatal(err)
	}
	var c dns.Client
	s := dns.Server{
		Net:  "udp",
		Addr: _listen,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			if filter(block, &r.Question); len(r.Question) == 0 {
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
