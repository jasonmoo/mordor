package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	host    = flag.String("host", "127.0.0.1", "host to scream at")
	workers = flag.Int("workers", 20, "number of concurrent requests in flight")
	timeout = flag.Duration("timeout", 300*time.Millisecond, "timeout for port request")

	start = flag.Int("start", 1025, "start of port range")
	end   = flag.Int("end", 65535, "start of port range")

	sema = make(chan struct{}, *workers)
)

func main() {

	flag.Parse()

	log.SetOutput(os.Stderr)

	// try full port range above root ports
	for i := *start; i < *end; i++ {
		sema <- struct{}{}
		go dial(i)
	}

}

func dial(port int) {
	defer func() { <-sema }()

	// log.Println("Dialing:", port)

	conn, err := net.DialTimeout("tcp", *host+":"+strconv.Itoa(port), *timeout)
	if err != nil {
		// log.Println(err)
		return
	}
	defer conn.Close()

	ok := make([]byte, 2)
	if _, err = conn.Read(ok); err != nil {
		log.Println(err)
		return
	}

	if string(ok) == "OK" {
		fmt.Println(port, "OK")
	} else {
		fmt.Println("received non-OK response:", string(ok), "from", port)
	}

}
