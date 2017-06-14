package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/howeyc/fsnotify"
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
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	go watcher.Watch(``)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				file := filepath.Base(ev.Name)
				if ev.IsModify() && file == BLOCK_FILE {
					block, err = read_blocks()
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}()
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
