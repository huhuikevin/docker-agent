package main

import (
	"flag"
	"fmt"

	"github.com/huhuikevin/docker-agent/services/agent"
)

func main() {
	fmt.Println("StartAgentServer")
	config := flag.String("config", "/etc/jwaoo/docker-agent.yaml", "config file")
	flag.Parse()
	agent.StartAgentWithConfigFile(*config)
}
