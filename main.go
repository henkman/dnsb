package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/henkman/trie"
	"github.com/judwhite/go-svc"
	"github.com/miekg/dns"
)

func main() {
	var opts struct {
		Resolve string `json:"resolve"`
		Listen  string `json:"listen"`
	}
	{
		exe, err := os.Executable()
		if err != nil {
			panic(err)
		}
		dir := filepath.Dir(exe)
		fd, err := os.Open(filepath.Join(dir, "dnsb.json"))
		if err != nil {
			panic(err)
		}
		err = json.NewDecoder(fd).Decode(&opts)
		fd.Close()
		if err != nil {
			panic(err)
		}
	}
	if opts.Resolve == "" {
		fmt.Println("resolve server not specified")
		return
	}
	s := Server{
		Listen:  opts.Listen,
		Resolve: opts.Resolve,
	}
	if err := svc.Run(&s); err != nil {
		panic(err)
	}
}

type Server struct {
	s       dns.Server
	Listen  string
	Resolve string
}

func (s *Server) Init(env svc.Environment) error {
	return nil
}

func (s *Server) Start() error {
	var bt trie.Trie
	if err := fillBlockTrie(&bt); err != nil {
		return err
	}
	go func() {
		var c dns.Client
		s.s = dns.Server{
			Net:  "udp",
			Addr: s.Listen,
			Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
				if filter(&bt, &r.Question); len(r.Question) == 0 {
					w.WriteMsg(r)
					return
				}
				in, _, err := c.Exchange(r, s.Resolve)
				if err != nil {
					dns.HandleFailed(w, r)
					return
				}
				w.WriteMsg(in)
			}),
		}
		s.s.ListenAndServe()
	}()
	return nil
}

func (s *Server) Stop() error {
	s.s.Shutdown()
	return nil
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
	fd, err := os.Open("dnsb.block")
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
