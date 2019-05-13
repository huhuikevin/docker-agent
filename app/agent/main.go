package main

import (
	"flag"
	"fmt"

	"jwaoo.com/services/agent"
)

func main() {
	fmt.Println("StartAgentServer")
	config := flag.String("config", "/etc/jwaoo/agent.yaml", "config file")
	flag.Parse()
	agent.StartAgentWithConfigFile(*config)
}
