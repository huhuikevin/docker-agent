package agent

import (
	"github.com/huhuikevin/docker-agent/uitls"
)

// proxy:
//   register: http://192.168.10.111:7000/api/v1/register
//   services: [oauth,common-msg,common-file]
//   beatheart: 2
// logs:
//   path: /data/server/agent
// docker:
//   cloud: ali
//   reposity: registry-vpc.cn-hongkong.aliyuncs.com
//   username: 'kevin@1734249857609980'
//   password: 'Reg&0928'
type docker struct {
	Cloud    string `json:"cloud,omitempty" yaml:"cloud,omitempty"`
	Regin    string `json:"regin,omitempty" yaml:"regin,omitempty"`
	Reposity string `json:"reposity,omitempty" yaml:"reposity,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

type logcfg struct {
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
	Level string `json:"level,omitempty" yaml:"level,omitempty"`
}

type proxy struct {
	Server    string   `json:"server,omitempty" yaml:"server,omitempty"`
	Keepalive string   `json:"keepalive,omitempty" yaml:"keepalive,omitempty"`
	Register  string   `json:"register,omitempty" yaml:"register,omitempty"`
	Services  []string `json:"services,omitempty" yaml:"services,omitempty"`
	Beatheart int      `json:"beatheart,omitempty" yaml:"beatheart,omitempty"`
}
type agentConfig struct {
	Proxy  proxy  `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	Logs   logcfg `json:"logs,omitempty" yaml:"logs,omitempty"`
	Docker docker `json:"docker,omitempty" yaml:"docker,omitempty"`
	Port   int16  `json:"port,omitempty" yaml:"port,omitempty"`
}

var agentConfigration = agentConfig{}

//LoadYAMLConfig load yaml config
func LoadYAMLConfig(config string) error {
	return uitls.LoadYAML(config, &agentConfigration)
}
