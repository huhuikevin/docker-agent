package proxy

import (
	"fmt"
	"sync"
	"time"

	"github.com/huhuikevin/docker-agent/common"
	"github.com/huhuikevin/docker-agent/logger"
	"github.com/huhuikevin/docker-agent/uitls"
	uuid "github.com/satori/go.uuid"
)

//Agent agent的配置
type Agent struct {
	Host          HostAddr
	Servers       []string
	QuitCh        chan int
	WaitQuitCh    chan int
	Stopping      bool
	Lock          *sync.Mutex
	Timer         *time.Ticker
	KeepTimeStamp int64
	UUID          string
}

const (
	startPath  = "/api/v1/start/"
	statusPath = "/api/v1/status/"
	stopPath   = "/api/v1/stop/"
)

var agents = make(map[HostAddr]*Agent, 0)
var agentLock = new(sync.Mutex)

//GetAgentsByServerName 获取可以运行serveice的agent
func GetAgentsByServerName(name string) []*Agent {
	agentLock.Lock()
	defer agentLock.Unlock()
	result := make([]*Agent, 0)
	for _, value := range agents {
		for _, s := range value.Servers {
			if s == name {
				result = append(result, value)
			}
		}
	}
	return result
}

//GetAgentByHostIP 通过host ip获取agent
func GetAgentByHostIP(ip string) *Agent {
	agentLock.Lock()
	defer agentLock.Unlock()
	for _, value := range agents {
		if ip == value.Host.IP() {
			return value
		}
	}
	return nil
}

func getAgentByHost(host string) *Agent {
	agentLock.Lock()
	defer agentLock.Unlock()
	if value, ok := agents[HostAddr(host)]; ok {
		return value
	}
	return nil
}

//RecvKeepAlived 收到 keep alive消息
func RecvKeepAlived(host string) bool {
	agent := getAgentByHost(host)
	if agent == nil {
		return false
	}
	//收到keepalive消息，重启定时器
	agent.stopCheckAvlible()
	agent.StartCheckAvalible()
	return true
}

//AddAgent 加入一个agent
func AddAgent(agent *Agent) {
	agentLock.Lock()
	var exsited *Agent
	if val, ok := agents[agent.Host]; ok {
		exsited = val
		logger.Println("delete existed host ", val.Host)
		delete(agents, exsited.Host)
	}
	agents[agent.Host] = agent
	logger.Println("add agent:", agent.Host, "services=", agent.Servers)
	agentLock.Unlock()
	if exsited != nil {
		exsited.stopCheckAvlible()
	}

	agent.StartCheckAvalible()
}

//DeleteAgent 删除一个agent
func DeleteAgent(agent *Agent) {
	agentLock.Lock()
	defer agentLock.Unlock()
	if val, ok := agents[agent.Host]; ok {
		if val.equls(agent) { //有可能被新的agent覆盖，但是HOST一样
			logger.Println("delete ", agent.Host)
			delete(agents, agent.Host)
			go agent.stopCheckAvlible()
		}
	}
}

//IsAgentInTable 判断agent是否已经添加到本地的table中
func IsAgentInTable(agent *Agent) bool {
	return IsHostAddrInTable(agent.Host)
}

//IsHostAddrInTable host地址代表的agent是否在本地的table中
func IsHostAddrInTable(host HostAddr) bool {
	agentLock.Lock()
	defer agentLock.Unlock()
	_, ok := agents[host]
	return ok
}

//NewAgent get an new agent from host and servers runing on it
func NewAgent(host string, servers []string) *Agent {
	agent := &Agent{Host: HostAddr(host), Servers: servers}
	agent.Lock = new(sync.Mutex)
	agent.QuitCh = make(chan int)
	agent.WaitQuitCh = make(chan int)
	uuid, err := uuid.NewV4()
	if err != nil {
		logger.Println("get UUID error:", err)
		return nil
	}
	agent.UUID = uuid.String()
	return agent
}

//newResult 封装common.NewResult，加入host
func (agent *Agent) newResult(errorcode common.ErrorCode) *common.RequstResult {
	result := common.NewResult(errorcode)
	result.Host = agent.Host.IP()
	return result
}

func (agent *Agent) equls(other *Agent) bool {
	if agent == other && agent.UUID == other.UUID {
		return true
	}
	return false
}

//GetServiceStatusByName 回去service的状态信息
func (agent *Agent) GetServiceStatusByName(name string) *common.RequstResult {
	if !agent.serviceCanRuning(name) {
		return agent.newResult(common.NotAllowedRuningOnHost)
	}
	agentURL := "http://" + agent.Host.String() + statusPath + name
	result := common.DoHTTPGet(agentURL)
	result.Host = agent.Host.IP()
	return result
}

//CheckServiceHealthByName 检查agent运行的service的监控状态
func (agent *Agent) CheckServiceHealthByName(name string, params *common.ServerParams) *common.RequstResult {
	if !agent.serviceCanRuning(name) {
		return agent.newResult(common.NotAllowedRuningOnHost)
	}
	serviceURL := fmt.Sprintf("http://%s:%d/%s", agent.Host.IP(), params.CheckHealth.Port, params.CheckHealth.Path)
	result, err := uitls.DoHTTPGet(serviceURL)
	if err != nil {
		return agent.newResult(common.LocalNetworkError)
	}
	if result.StatusCode != params.CheckHealth.Code {
		result1 := agent.newResult(common.ServiceNotHealth)
		msg := fmt.Sprintf("Error, statusCode = %d", result.StatusCode)
		msg = "[" + msg + "]"
		result1.Data = msg
		return result1
	}
	result1 := agent.newResult(common.Success)
	msg := "OK"
	msg = "[" + msg + "]"
	result1.Data = msg
	return result1
}

//StartServiceByName 启动可以运行在agent上的一个服务
func (agent *Agent) StartServiceByName(name string, params *common.ServerParams) *common.RequstResult {
	if !agent.serviceCanRuning(name) {
		return agent.newResult(common.NotAllowedRuningOnHost)
	}
	agentURL := "http://" + agent.Host.String() + startPath + name
	result := common.DoHTTPPostJSON(agentURL, params)
	result.Host = agent.Host.IP()
	return result
}

//StopServiceByName stop the service
func (agent *Agent) StopServiceByName(name string, params *common.ServerParams) *common.RequstResult {
	agentURL := "http://" + agent.Host.String() + stopPath + name
	result := common.DoHTTPPostJSON(agentURL, params)
	result.Host = agent.Host.IP()
	return result
}

func (agent *Agent) serviceCanRuning(name string) bool {
	for _, val := range agent.Servers {
		if name == val {
			return true
		}
	}
	return false
}

//StartCheckAvalible 开始检查agent是否可用
func (agent *Agent) StartCheckAvalible() {
	agent.KeepTimeStamp = time.Now().Unix()
	go agent.checkAvalible()
}

func (agent *Agent) checkAvalible() {
	agent.Timer = time.NewTicker(time.Duration(proxyConfigration.Checkagent.Interval) * time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-agent.Timer.C:
				time := time.Now().Unix()
				if (time - agent.KeepTimeStamp) >= int64(proxyConfigration.Checkagent.Interval) {
					agent.processError()
				}
			case <-agent.QuitCh:
				//logger.Println("Recv Quit signal")
				agent.Timer.Stop()
				return
			}
		}
	}()
	wg.Wait()
	agent.WaitQuitCh <- 1
}

func (agent *Agent) processError() {
	DeleteAgent(agent)
}

func (agent *Agent) stopCheckAvlible() {
	agent.QuitCh <- 1 //stop check routine
	select {
	case <-agent.WaitQuitCh:
		return
	}
}
