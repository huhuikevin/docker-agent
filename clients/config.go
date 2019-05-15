package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/huhuikevin/docker-agent/logger"

	"github.com/huhuikevin/docker-agent/uitls"
)

//HealthCheck define the service's healcheck
type HealthCheck struct {
	Path    string `json:"path,omitempty" yaml:"path,omitempty"`
	Code    int    `json:"code,omitempty" yaml:"code,omitempty"`
	Timeout int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

//Ulimits define the service's ulimits of the linux system
type Ulimits struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Soft int    `json:"softLimit,omitempty" yaml:"softLimit,omitempty"`
	Hard int    `json:"hardLimit,omitempty" yaml:"hardLimit,omitempty"`
}

//ServerConfig read from json or yaml config files to define the server properties
type ServerConfig struct {
	Image        string      `json:"image,omitempty" yaml:"image,omitempty"`
	DockerName   string      `json:"container,omitempty" yaml:"container,omitempty"`
	Hosts        []string    `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	PortMap      []string    `json:"ports,omitempty" yaml:"ports,omitempty"`
	DubboPortMap []string    `json:"dubbo_ports,omitempty" yaml:"dubbo_ports,omitempty"`
	NetworkMode  string      `json:"network,omitempty" yaml:"network,omitempty"`
	Volume       []string    `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Memory       string      `json:"memory,omitempty" yaml:"memory,omitempty"`
	CheckHealth  HealthCheck `json:"checkHealth,omitempty" yaml:"checkHealth,omitempty"`
	Ulimits      []Ulimits   `json:"ulimits,omitempty" yaml:"ulimits,omitempty"`
	JavaOpts     string      `json:"java_opts,omitempty" yaml:"java_opts,omitempty"`
	ExtEnv       []string    `json:"environment,omitempty" yaml:"environment,omitempty"`
	ExtraHosts   []string    `json:"extra_hosts,omitempty" yaml:"extra_hosts,omitempty"`
	RunCmd       []string    `json:"runCmd,omitempty" yaml:"runCmd,omitempty"`
}

//RunningConfig read from json or yaml config files
//Repository docker hub地址，如果没有指定使用agent配置的地址
//ProxyHTTP docker管理的地址，所有能运行docker的主机都会注册到这个server上
//ComposeServer 需要多个docker一起工作的server，比如zk，mq，fdfs等
type RunningConfig struct {
	Repository    string                  `json:"repository,omitempty" yaml:"repository,omitempty"`
	ProxyHTTP     string                  `json:"docker_agent,omitempty" yaml:"docker_agent,omitempty"`
	ComposeServer string                  `json:"compose_server,omitempty" yaml:"compose_server,omitempty"`
	Servers       map[string]ServerConfig `json:"services,omitempty" yaml:"services,omitempty"`
}

var config = RunningConfig{}

//LoadConfigfileByServer load config files,and return the RunningConfig
func LoadConfigfileByServer(path string, server string) *RunningConfig {
	cfgfiles := getConfigFile(path)
	for _, file := range cfgfiles {
		temp := RunningConfig{}
		//loadJSON(file, &temp)
		//fmt.Println(file)
		loadYAML(file, &temp)
		//fmt.Println(temp)
		if isServerConfig(server, temp) {
			config = temp
			fmt.Println("found config file:", file)
			break
		}
	}
	if len(config.Servers) == 0 {
		fmt.Println("can not found config file for ", server)
		os.Exit(1)
	}
	return &config
}

func isServerConfig(name string, config RunningConfig) bool {
	if config.ComposeServer == name {
		return true
	}
	for k := range config.Servers {
		if name == k {
			return true
		}
	}
	return false
}

func loadYAML(filename string, v interface{}) error {
	return uitls.LoadYAML(filename, v)
}

func loadJSON(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func getConfigFile(path string) []string {
	configs := make([]string, 0, 1)
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		//fmt.Println(f.Name())
		if strings.Contains(f.Name(), ".yaml") {
			file := path + "/" + f.Name()
			configs = append(configs, file)
		}
	}
	return configs
}

//GetProxyHTTPURL get docker agent http address
func GetProxyHTTPURL() string {
	return config.ProxyHTTP
}

//BuildParamsByServerName ...
func BuildParamsByServerName(config RunningConfig, name string, tag string) []map[string]interface{} {
	params := make([]map[string]interface{}, 0, 1)
	for k, v := range config.Servers {
		if k == name || config.ComposeServer == name { //有可能同一个服务跑多个docker，比如zk，fdfs，mq等
			parts := strings.Split(v.Image, ":")
			if tag == "" {
				if len(parts) == 1 {
					v.Image = parts[0] + ":" + "latest"
					logger.Println("use tag latest")
				}
			} else {
				v.Image = parts[0] + ":" + tag
			}
			if len(v.Hosts) > 0 { //指定host
				for _, host := range v.Hosts { //一个服务在多台HOST上跑
					p := buildParams(name, v, host)
					params = append(params, p)
				}
			} else {
				p := buildParams(name, v, "")
				params = append(params, p)
			}
		}
	}
	return params
}

func buildDockerImagePath(sconfig ServerConfig) string {

	if config.Repository != "" {
		return fmt.Sprintf("%s/%s", config.Repository, sconfig.Image)
	}
	return sconfig.Image
}

func string2int(s string) int {
	m, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return int(m)

}

func buildParams(name string, config ServerConfig, host string) map[string]interface{} {
	params := make(map[string]interface{})

	params["Env"] = buildEnv(config, host)
	params["Volume"] = buildVolume(config)
	params["Port"] = buildPortMaps(config)
	params["NetworkMode"] = config.NetworkMode
	memory := buildMemory(config)
	if memory != 0 {
		params["Memory"] = memory
		params["MemorySwap"] = memory
	}
	if len(config.ExtraHosts) > 0 {
		params["extraHosts"] = config.ExtraHosts
	}

	if len(config.RunCmd) > 0 {
		params["cmd"] = config.RunCmd
	}
	params["Ulimits"] = buildUlimits(config)

	serverConfig := make(map[string]interface{})
	serverConfig["config"] = params
	serverConfig["credential"] = "aaaaaaaaaa3$55"
	serverConfig["image"] = buildDockerImagePath(config)
	if config.CheckHealth.Path != "" {
		checkHealth := make(map[string]interface{})
		checkHealth["methend"] = "http"
		checkHealth["path"] = config.CheckHealth.Path
		parts := strings.Split(config.PortMap[0], ":")
		checkHealth["port"] = string2int(parts[1])
		checkHealth["code"] = config.CheckHealth.Code
		serverConfig["checkHealth"] = checkHealth
	}
	if host != "" {
		serverConfig["host"] = host
	}
	if config.DockerName == "" {
		config.DockerName = name
	}
	serverConfig["containerName"] = config.DockerName
	return serverConfig
}

func buildEnv(config ServerConfig, host string) []string {
	env := make([]string, len(config.ExtEnv))
	copy(env, config.ExtEnv)
	//如果没有指定host，用$HOST_IP替代，agent上会把$HOST_IP替换为本机真实的ip
	if host == "" {
		host = "$HOST_IP"
	}
	if config.JavaOpts != "" {
		env = append(env, config.JavaOpts)
	}
	penv := fmt.Sprintf("EXTERNAL_IP=%s", host)
	env = append(env, penv)
	if len(config.PortMap) > 0 {
		//fmt.Println("portMap=", config.PortMap)
		parts := strings.Split(config.PortMap[0], ":")
		penv = fmt.Sprintf("APP_PORT=%d", string2int(parts[0]))
		env = append(env, penv)
		penv = fmt.Sprintf("HOST_PORT=%d", string2int(parts[1]))
		env = append(env, penv)
		penv = fmt.Sprintf("PORT_%d=%d", string2int(parts[0]), string2int(parts[1]))
		env = append(env, penv)

		if len(config.PortMap) == 2 {
			parts := strings.Split(config.PortMap[1], ":")
			penv := fmt.Sprintf("APP_PORT2=%d", string2int(parts[0]))
			env = append(env, penv)
			penv = fmt.Sprintf("HOST_PORT2=%d", string2int(parts[1]))
			env = append(env, penv)
		}
	}
	if len(config.DubboPortMap) > 0 {
		parts := strings.Split(config.DubboPortMap[0], ":")
		penv := fmt.Sprintf("DUBBO_PORT_TO_REGISTRY=%d", string2int(parts[1]))
		env = append(env, penv)
		penv = fmt.Sprintf("DUBBO_IP_TO_REGISTRY=%s", host)
		env = append(env, penv)
	}
	penv = fmt.Sprintf("HOST_IP=%s", host)
	env = append(env, penv)
	return env
}

func buildVolume(config ServerConfig) []string {
	return config.Volume
}

func buildPortMaps(config ServerConfig) map[string]string {
	portMaps := make(map[string]string)
	for _, v := range config.PortMap {
		parts := strings.Split(v, ":")
		cport := parts[0]
		hport := parts[1]
		portMaps[cport] = hport
	}
	if len(config.DubboPortMap) > 0 {
		parts := strings.Split(config.DubboPortMap[0], ":")
		cport := parts[0]
		hport := parts[1]
		portMaps[cport] = hport
	}
	return portMaps
}

func buildMemory(config ServerConfig) int64 {
	if config.Memory == "" {
		return 0
	}
	data := []byte(config.Memory)
	var mult int64 = 1
	last := data[len(data)-1]
	if last == 'K' || last == 'k' {
		mult = 1024
	} else if last == 'M' || last == 'm' {
		mult = 1024 * 1024
	} else if last == 'G' || last == 'g' {
		mult = 1024 * 1024 * 1024
	} else {
		m, err := strconv.ParseInt(config.Memory, 10, 64)
		if err != nil {
			panic(err)
		}
		return m
	}
	mstr := string(data[0 : len(data)-1])
	m, err := strconv.ParseInt(mstr, 10, 64)
	if err != nil {
		panic(err)
	}
	return m * mult
}

func buildNetworkMode(config ServerConfig) string {
	return config.NetworkMode
}

func buildHostName(config ServerConfig) string {
	return ""
}

func buildUlimits(config ServerConfig) []Ulimits {
	return config.Ulimits
}
