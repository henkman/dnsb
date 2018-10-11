package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/henkman/trie"
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

func main() {
	if _resolve == "" {
		flag.Usage()
		return
	}
	var bt trie.Trie
	if err := fillBlockTrie(&bt); err != nil {
		log.Fatal(err)
	}
	var c dns.Client
	s := dns.Server{
		Net:  "udp",
		Addr: _listen,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			if filter(&bt, &r.Question); len(r.Question) == 0 {
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

func filter(bt *trie.Trie, qs *[]dns.Question) {
	i := 0
next:
	for i < len((*qs)) {
		n := (*qs)[i].Name
		n = n[:len(n)-1]
		e := bt.Lookup(reverse(n))
		if e != nil && (len(e.Children) == 0 || e.Leaf) {
			(*qs)[i] = (*qs)[len((*qs))-1]
			(*qs) = (*qs)[:len((*qs))-1]
			continue next
		}
		i++
	}
}

func fillBlockTrie(bt *trie.Trie) error {
	fd, err := os.Open(BLOCK_FILE)
	if err != nil {
		return err
	}
	bin := bufio.NewReader(fd)
	for {
		line, _ := bin.ReadString('\n')
		if line == "" {
			break
		}
		bt.Insert(reverse(strings.TrimSpace(line)))
	}
	fd.Close()
	return nil
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
