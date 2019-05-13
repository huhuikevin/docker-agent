package proxy

import (
	"jwaoo.com/uitls"
)

// port: 8000
// redis:
//   server: http://1.1.1.1:6397
//   password: ""
//   redisDB: 1

type redis struct {
	Server   string `json:"server,omitempty" yaml:"server,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	RedisDB  int    `json:"redisDB,omitempty" yaml:"redisDB,omitempty"`
}

type logcfg struct {
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
	Level string `json:"level,omitempty" yaml:"level,omitempty"`
}

type checkagent struct {
	Interval   int `json:"interval,omitempty" yaml:"interval,omitempty"`
	Thrldcount int `json:"thrldcount,omitempty" yaml:"thrldcount,omitempty"`
}
type proxyConfig struct {
	Redis      redis      `json:"redis,omitempty" yaml:"redis,omitempty"`
	Logs       logcfg     `json:"logs,omitempty" yaml:"logs,omitempty"`
	Port       int16      `json:"port,omitempty" yaml:"port,omitempty"`
	Checkagent checkagent `json:"checkagent,omitempty" yaml:"checkagent,omitempty"`
}

var proxyConfigration = proxyConfig{}

//LoadYAMLConfig load yaml config
func LoadYAMLConfig(config string) error {
	return uitls.LoadYAML(config, &proxyConfigration)
}
