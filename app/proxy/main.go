package main

import (
	"flag"
	"fmt"

	"jwaoo.com/services/proxy"
)

func main() {
	fmt.Println("StartProxyServer")
	config := flag.String("config", "/etc/jwaoo/proxy.yaml", "config yaml")
	flag.Parse()
	proxy.StartProxyWithConfigFile(*config)
}
