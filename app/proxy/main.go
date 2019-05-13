package main

import (
	"flag"
	"fmt"

	"github.com/huhuikevin/docker-agent/services/proxy"
)

func main() {
	fmt.Println("StartProxyServer")
	config := flag.String("config", "/etc/jwaoo/proxy.yaml", "config yaml")
	flag.Parse()
	proxy.StartProxyWithConfigFile(*config)
}
