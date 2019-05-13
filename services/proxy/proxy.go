package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huhuikevin/docker-agent/common"
	"github.com/huhuikevin/docker-agent/logger"
	log "github.com/huhuikevin/docker-agent/logger"
)

type serverParams common.ServerParams

//AgentPort the port of the agent
var AgentPort int16 = 8001

func httpProcessKeepalive(ctx *gin.Context) {
	params := ""
	if err := ctx.BindJSON(&params); err != nil {
		log.Println("bindjson error:", err.Error())
		result := common.NewResult(common.CannotMapToResult)
		result.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, result)
		return
	}

	if RecvKeepAlived(params) {
		ctx.JSON(http.StatusOK, common.NewResult(common.Success))
		return
	}
	ctx.JSON(http.StatusOK, common.NewResult(common.AgentHaveNotRegistedOnProxy))
}

func httpProcessRegister(ctx *gin.Context) {
	params := common.RegisterParams{}
	if err := ctx.BindJSON(&params); err != nil {
		log.Println("bindjson error:", err.Error())
		result := common.NewResult(common.CannotMapToResult)
		result.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, result)
		return
	}
	agent := NewAgent(params.Host, params.Services)
	AddAgent(agent)
	//AddAgentServers(params.Services, params.Host)
	ctx.JSON(http.StatusOK, common.NewResult(common.Success))
}

func postStartCmdToAgent(server string, params *common.ServerParams) []*common.RequstResult {
	messages := make([]*common.RequstResult, 0)
	//如果参数中指定要在哪个IP对应的host上运行，找到这个IP对应的agent，这个一般在启动基础服务的时候需要
	if params.Host != "" {
		agent := GetAgentByHostIP(params.Host)
		result := agent.StartServiceByName(server, params)
		messages = append(messages, result)
		return messages
	}
	agents := GetAgentsByServerName(server)
	log.Println("agent = ", agents)
	if len(agents) == 0 {
		errResult := common.NewResult(common.CannotFoundAvalibleHost)
		messages = append(messages, errResult)
		return messages
	}
	for _, agent := range agents {
		result := agent.StartServiceByName(server, params)
		messages = append(messages, result)
	}
	return messages
}

func postCheckhealCmdToAgend(server string, params *common.ServerParams) []*common.RequstResult {
	messages := make([]*common.RequstResult, 0)
	//如果参数中指定要在哪个IP对应的host上运行，找到这个IP对应的agent，这个一般在启动基础服务的时候需要
	if params.Host != "" {
		agent := GetAgentByHostIP(params.Host)
		result := agent.CheckServiceHealthByName(server, params)
		messages = append(messages, result)
		return messages
	}
	agents := GetAgentsByServerName(server)
	if agents == nil {
		errResult := common.NewResult(common.CannotFoundAvalibleHost)
		messages = append(messages, errResult)
		return messages
	}
	for _, agent := range agents {
		result := agent.CheckServiceHealthByName(server, params)
		messages = append(messages, result)
	}
	return messages
}

func postGetStatusCmdToAgend(server string) []*common.RequstResult {
	messages := make([]*common.RequstResult, 0)
	agents := GetAgentsByServerName(server)
	if agents == nil {
		errResult := common.NewResult(common.CannotFoundAvalibleHost)
		messages = append(messages, errResult)
		return messages
	}
	for _, agent := range agents {
		result := agent.GetServiceStatusByName(server)
		messages = append(messages, result)
	}
	return messages
}

func httpTransfer(ctx *gin.Context) {
	server := ctx.Param("server")
	if ctx.Request.Method == "POST" {
		params := common.ServerParams{}
		if err := ctx.BindJSON(&params); err == nil {
			if params.Credential != common.Credential {
				httpResult := common.NewResult(common.CredentialError)
				ctx.JSON(http.StatusBadRequest, httpResult)
				logger.Println("认证失败")
				return
			}
		} else {
			httpResult := common.NewResult(common.ParamsNotValide)
			ctx.JSON(http.StatusBadRequest, httpResult)
			logger.Println("JSON bind失败")
			return
		}
		if strings.Contains(ctx.Request.URL.Path, "/start/") {
			results := postStartCmdToAgent(server, &params)
			statusCode := http.StatusOK
			for _, result := range results {
				if result.Code != common.Success {
					statusCode = http.StatusBadRequest
					break
				}
			}
			if statusCode != http.StatusOK {
				resultData := common.NewResult(common.StartServiceError)
				resultData.Data = results
				ctx.JSON(statusCode, resultData)
				return
			}
			resultData := common.NewResult(common.Success)
			resultData.Data = results
			ctx.JSON(statusCode, resultData)
		} else if strings.Contains(ctx.Request.URL.Path, "/check/") {
			results := postCheckhealCmdToAgend(server, &params)
			statusCode := http.StatusOK
			for _, result := range results {
				if result.Code != common.Success {
					statusCode = http.StatusBadRequest
					break
				}
			}
			if statusCode != http.StatusOK {
				resultData := common.NewResult(common.CheckServiceHealthError)
				resultData.Data = results
				ctx.JSON(statusCode, resultData)
				return
			}
			resultData := common.NewResult(common.Success)
			resultData.Data = results
			ctx.JSON(statusCode, resultData)
		}
	} else if strings.Contains(ctx.Request.URL.Path, "/status/") {
		results := postGetStatusCmdToAgend(server)
		statusCode := http.StatusOK
		for _, result := range results {
			if result.Code != common.Success {
				statusCode = http.StatusBadRequest
				break
			}
		}
		if statusCode != http.StatusOK {
			resultData := common.NewResult(common.GetServiceStatusError)
			resultData.Data = results
			ctx.JSON(statusCode, resultData)
			return
		}
		resultData := common.NewResult(common.Success)
		resultData.Data = results
		ctx.JSON(statusCode, resultData)
	}
}

//StartProxyWithConfigFile start proxy using config file
func StartProxyWithConfigFile(config string) {
	LoadYAMLConfig(config)
	InitRouteTalbe()
	StartProxyServer(proxyConfigration.Port)
}

//StartProxyServer start proxy server with the port
func StartProxyServer(port int16) {
	engine := gin.Default()
	group := engine.Group("/api/v1")
	group.POST("/start/:server", httpTransfer)
	group.POST("/check/:server", httpTransfer)
	group.GET("/status/:server", httpTransfer)
	group.POST("/register", httpProcessRegister)
	group.POST("/keepalive", httpProcessKeepalive)
	//AgentPort = agentPort
	sport := fmt.Sprintf(":%d", port)
	err := engine.Run(sport)
	if err != nil {
		fmt.Println(err)
	}
}
