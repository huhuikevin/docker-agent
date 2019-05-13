package common

import (
	client "github.com/huhuikevin/docker-agent/dockerclient"
)

//Credential credential of the agent
var Credential = "aaaaaaaaaa3$55"

//ServerParams tell the agent how can start the docker container
type ServerParams struct {
	Credential string `json:"credential,omitempty" yaml:"credential,omitempty" toml:"credential,omitempty"`
	//server     string                 `json:"server,omitempty" yaml:"server,omitempty" toml:"server,omitempty"`
	Image         string                 `json:"image,omitempty" image:"server,omitempty" image:"server,omitempty"`
	Config        client.ContainerConfig `json:"config,omitempty" yaml:"config,omitempty" toml:"config,omitempty"`
	CheckHealth   checkHealth            `json:"CheckHealth,omitempty" yaml:"CheckHealth,omitempty"`
	Host          string                 `json:"Host,omitempty" yaml:"Host,omitempty"`
	ContainerName string                 `json:"containerName,omitempty" yaml:"containerName,omitempty"`
}

type checkHealth struct {
	Methend string `json:"Methend,omitempty" yaml:"Methend,omitempty"`
	Path    string `json:"Path,omitempty" yaml:"Path,omitempty"`
	Port    int16  `json:"Port,omitempty" yaml:"Port,omitempty"`
	Code    int    `json:"Code,omitempty" yaml:"Code,omitempty"`
}

// type serverConfig struct {
// 	Name  string   `json:"Name,omitempty" yaml:"Name,omitempty"`
// 	Hosts []string `json:"Hosts,omitempty" yaml:"Hosts,omitempty"`
// }

//RegisterParams agent register to proxy
type RegisterParams struct {
	Host     string   `json:"host,omitempty" yaml:"host,omitempty"`
	Services []string `json:"services,omitempty" yaml:"services,omitempty"`
}
