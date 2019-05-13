package proxy

import (
	"strings"
	"sync"

	"encoding/json"

	"jwaoo.com/logger"
	rclient "jwaoo.com/redis"
)

//HostAddr ip:port or ip
type HostAddr string

var tableLock = new(sync.Mutex)

type hostTableType map[string][]HostAddr

var hostTable = hostTableType{}

var redisClient *rclient.Client

func saveRoute(names []string, host string) {
	if redisClient == nil {
		return
	}
	data, err := json.Marshal(names)
	if err != nil {
		logger.Println("json marshal error:", err.Error())
		return
	}
	redisClient.SetValue(host, data)
}

//InitRouteTalbe 初始化route table表，启动的时候执行，从redis读取agent注册上来的信息
func InitRouteTalbe() {
	initRedis()
	// keys, err := redisClient.GetAllKeys()
	// if err != nil {
	// 	logger.Println("redis: get all keys error")
	// 	return
	// }
	// logger.Println("Read running config from redis, hosts = ", keys)
	// for _, host := range keys {
	// 	datas, err := redisClient.GetValue(host)
	// 	if err == nil {
	// 		servers := make([]string, 0)
	// 		err = json.Unmarshal(datas, &servers)
	// 		if err == nil {
	// 			AddAgentServers(servers, host)
	// 		} else {
	// 			logger.Println("json unmarshal error:", err.Error())
	// 		}
	// 	} else {
	// 		logger.Println("get redis value error:", err.Error())
	// 	}
	// }
	// logger.Println("Read running config from redis end")
}

//initRedis 初始化redis
func initRedis() error {
	if proxyConfigration.Redis.Server == "" {
		logger.Println("do'nt use redis")
		return nil
	}
	if client, err := rclient.GetDB(proxyConfigration.Redis.Server, proxyConfigration.Redis.Password, proxyConfigration.Redis.RedisDB); err == nil {
		redisClient = client
	} else {
		logger.Println("init redis error:", err.Error())
		return err
	}
	return nil
}

//AddAgentServer add server,host pair to the route table
func AddAgentServer(name, host string) {
	defer tableLock.Unlock()
	tableLock.Lock()
	hosts := []HostAddr(nil)
	if val, ok := hostTable[name]; ok {
		hosts = val
	} else {
		hosts = make([]HostAddr, 0)
	}
	logger.Println("add server:", name, "on host:", host)
	hosts = append(hosts, HostAddr(host))
	hostTable[name] = hosts
}

//AddAgentServers 一般用在一台host上线，可能跑多个service
func AddAgentServers(names []string, host string) {
	for _, v := range names {
		AddAgentServer(v, host)
	}
	saveRoute(names, host)
}

//DelAgentServer delete server and ip:port
func DelAgentServer(name, host string) {
	defer tableLock.Unlock()
	tableLock.Lock()
	hosts := make([]HostAddr, 0)
	if val, ok := hostTable[name]; ok {
		for _, v := range val {
			if v.String() != host {
				hosts = append(hosts, v)
			}
		}
	}
	hostTable[name] = hosts
}

//DelAgentServers 一般用在一台host下线，可能跑多个service
func DelAgentServers(names []string, host string) {
	for _, v := range names {
		DelAgentServer(v, host)
	}
}

//GetAgentsByServerName 获取server对应的多个host
// func GetAgentsByServerName(name string) []HostAddr {
// 	defer tableLock.Unlock()
// 	tableLock.Lock()
// 	allhost := make([]HostAddr, 0)
// 	//ALL 代表所有的服务都可以在ALL对应的host上运行
// 	if val, ok := hostTable["ALL"]; ok {
// 		allhost = val
// 	}
// 	if val, ok := hostTable[name]; ok {
// 		hosts := append(val, allhost...)
// 		return hosts
// 	}
// 	return nil
// }

//GetAgentByGivenHost 通过server和host查找agent
// func GetAgentByGivenHost(name, host string) []HostAddr {
// 	hosts := GetAgentsByServerName(name)
// 	if hosts == nil || len(hosts) == 0 {
// 		return nil
// 	}
// 	for _, one := range hosts {
// 		if one.IP() == host {
// 			return []HostAddr{one}
// 		}
// 	}
// 	return nil
// }

//IP ip address of hostadd
func (me HostAddr) IP() string {
	parts := strings.Split(me.String(), ":")
	return parts[0]
}

//PORT port address of hostadd
func (me HostAddr) PORT() string {
	parts := strings.Split(me.String(), ":")
	if len(parts) <= 1 {
		return ""
	}
	return parts[1]
}

//String port address of hostadd
func (me HostAddr) String() string {
	return string(me)
}
