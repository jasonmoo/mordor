package main

import (
	"flag"
	"log"
	"net"
)

var (
	// expects iptables rule:
	// iptables -t nat -A PREROUTING -p tcp --dport 1025:65535 -j REDIRECT --to-ports 10000
	//
	host = flag.String("host", ":10000", "host to listen on")
)

func main() {

	ln, err := net.Listen("tcp", *host)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = conn.Write([]byte("OK"))
		if err != nil {
			log.Println(err)
		}
		conn.Close()
	}

}
