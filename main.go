package main

import (
	"flag"
	"forward/proxy"
	"log"
)

func main() {
	var isServer, isClient bool
	var lnetwork, rnetwork, laddress, raddress, secret string

	flag.BoolVar(&isServer, "s", false, "run as server")
	flag.BoolVar(&isClient, "c", false, "run as client")
	flag.StringVar(&lnetwork, "lnet", "tcp", "listen network")
	flag.StringVar(&rnetwork, "rnet", "tcp", "remote network")
	flag.StringVar(&laddress, "laddr", "", "listen address")
	flag.StringVar(&raddress, "raddr", "", "remote address")
	flag.StringVar(&secret, "secret", "", "secret")
	flag.Parse()

	if !isServer && !isClient || isServer && isClient || len(laddress) == 0 || len(raddress) == 0 {
		flag.PrintDefaults()
		return
	}

	if isServer {
		fs, err := proxy.NewForwardServer(lnetwork, rnetwork, laddress, raddress, []byte(secret))
		if err != nil {
			log.Fatalln(err)
		}
		fs.Run()
	} else if isClient {
		fc, err := proxy.NewForwardClient(lnetwork, rnetwork, laddress, raddress, []byte(secret))
		if err != nil {
			log.Fatalln(err)
		}
		fc.Run()
	}
}
