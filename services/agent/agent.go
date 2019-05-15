package agent

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huhuikevin/docker-agent/common"
	"github.com/huhuikevin/docker-agent/logger"
	"github.com/huhuikevin/docker-agent/uitls"
)

type serverParams common.ServerParams

//myIP 本机ip地址，可以外网连接的，不是docker，loopback等地址
var myIP = ""

//newResult 封装common.NewResult，加入host
func newResult(errorcode common.ErrorCode) *common.RequstResult {
	result := common.NewResult(errorcode)
	result.Host = myIP
	return result
}

func httpAgentStatus(ctx *gin.Context) {
	httpResult := newResult(common.Success)
	ctx.JSON(http.StatusOK, httpResult)
}

func httpStatusHandler(ctx *gin.Context) {
	server := ctx.Param("server")
	logger.Println("GET server=", server)
	info := uitls.FindContainerInfoByName(server)
	if _, ok := info["ID"]; ok {
		parts := strings.Split(info["Image"], "/")
		httpResult := newResult(common.Success)
		image := parts[len(parts)-1]
		httpResult.Data = image
		ctx.JSON(http.StatusOK, httpResult)
	} else {
		httpResult := newResult(common.ServiceNotFound)
		ctx.JSON(http.StatusBadRequest, httpResult)
	}
}

func errorMsg(msg string, err error) string {
	return fmt.Sprintf("%s:%s", msg, err.Error())
}

func getDockerImage(name string) string {
	//如果name中已经包含了厂库地址，直接返回
	if strings.Contains(name, "/") {
		return name
	}
	return agentConfigration.Docker.Reposity + "/" + name
}

func processStartServer(server string, params serverParams, ctx *gin.Context) error {
	params.Image = getDockerImage(params.Image)
	logger.Println("Start Docker:", params.Image)
	logger.Println("Docker name:", params.ContainerName)
	if err := uitls.PullDockerImage(params.Image, agentConfigration.Docker.Username, agentConfigration.Docker.Password); err != nil {
		httpResult := newResult(common.PullImageError)
		httpResult.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, httpResult)
		logger.Println("拉取Image失败:", err.Error())
		return err
	}
	if err := uitls.StopContainerByName(params.ContainerName); err != nil {
		httpResult := newResult(common.StopContainerError)
		httpResult.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, httpResult)
		logger.Println("停止docker失败:", err.Error())
		return err
	}
	// //处理$HOSTNAME，$HOST_IP等需要替换的字符串，类似linux shell的变量
	env := params.Config.Env
	for index, val := range env {
		if strings.Contains(val, "$HOSTNAME") {
			hostname, err := os.Hostname()
			if err != nil {
				hostname = myIP
			}
			env[index] = strings.Replace(val, "$HOSTNAME", hostname, -1)
		}
		if strings.Contains(val, "$HOST_IP") {
			env[index] = strings.Replace(env[index], "$HOST_IP", myIP, -1)
		}
	}
	if _, err := uitls.StartDocker(params.Image, params.Config, params.ContainerName); err != nil {
		httpResult := newResult(common.StartContainerError)
		httpResult.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, httpResult)
		logger.Println("启动docker失败:", err.Error())
		return err
	}
	return nil
}

func httpStartHandler(ctx *gin.Context) {
	server := ctx.Param("server")
	params := serverParams{}

	if ctx.BindJSON(&params) == nil {
		if params.Credential != common.Credential {
			httpResult := newResult(common.CredentialError)
			ctx.JSON(http.StatusBadRequest, httpResult)
			logger.Println("认证失败")
			return
		}
	} else {
		httpResult := newResult(common.ParamsNotValide)
		ctx.JSON(http.StatusBadRequest, httpResult)
		logger.Println("JSON bind失败")
		return
	}

	if processStartServer(server, params, ctx) == nil {
		info := uitls.FindContainerInfoByName(params.ContainerName)
		if _, ok := info["ID"]; ok {
			httpResult := newResult(common.Success)
			httpResult.Data = info
			ctx.JSON(http.StatusOK, httpResult)
		} else {
			logger.Println("Docker name:", params.ContainerName)
			httpResult := newResult(common.StartContainerError)
			parts := strings.Split(params.Image, "/")
			msg := fmt.Sprintf("*Docker Started, Check Again Failt -- %s", parts[len(parts)-1])
			httpResult.Data = msg
			ctx.JSON(http.StatusBadRequest, httpResult)
		}
	}
}

func runRegisterRoutine(quit <-chan int) {
	ticker := time.NewTicker(time.Duration(agentConfigration.Proxy.Beatheart) * time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	register := true //ture 需要register，false 需要keepalive
	go func() {
		defer wg.Done()
		logger.Println("runRegisterRoutine")
		for {
			select {
			case <-ticker.C:
				if register {
					if err := doRegister(); err == nil {
						logger.Println("runRegisterRoutine done")
						register = false
					}
				} else {
					if err := doKeeperAlive(); err == nil {
						logger.Println("doKeeperAlive done")
						register = false
					} else {
						logger.Println("switch to  doRegister")
						register = true
					}
				}
			case <-quit:
				logger.Println("work well exist.")
				ticker.Stop()
				return
			}
		}
	}()
	wg.Wait()
}

func ip() string {
	ip := os.Getenv("HOST_IP")
	if ip == "" {
		ip = uitls.GetLocalIPByAccessHTTPServer(agentConfigration.Proxy.Server)
	}
	port := agentConfigration.Port
	ports := os.Getenv("APP_PORT")
	if ports != "" {
		_port, err := strconv.Atoi(ports)
		if err != nil {
			return ""
		}
		port = int16(_port)
	}

	return fmt.Sprintf("%s:%d", ip, port)
}

func myIPPORT() (string, error) {
	host := myIP
	if host == "" {
		host = ip()
		myIP = host
	}
	if host == "" {
		return "", errors.New("get local ip error")
	}
	return host, nil
}

func doRegister() error {
	registerURL := agentConfigration.Proxy.Server + agentConfigration.Proxy.Register
	host, err := myIPPORT()
	if err != nil {
		return err
	}
	logger.Println("registerURL=", registerURL)

	logger.Println("myip=", host)
	params := common.RegisterParams{Services: agentConfigration.Proxy.Services, Host: host}
	result := common.DoHTTPPostJSON(registerURL, params)
	if result.Code == common.Success {
		return nil
	}
	return errors.New("result.Message")
}

func doKeeperAlive() error {
	keepaliveURL := agentConfigration.Proxy.Server + agentConfigration.Proxy.Keepalive

	host, err := myIPPORT()
	if err != nil {
		return err
	}
	result := common.DoHTTPPostJSON(keepaliveURL, host)
	if result.Code == common.Success {
		return nil
	}
	return errors.New(result.Message)
}

//StartAgentWithConfigFile start the agent with yaml config
func StartAgentWithConfigFile(file string) {
	if err := LoadYAMLConfig(file); err != nil {
		logger.Println("load yaml config error, " + err.Error())
		return
	}
	if agentConfigration.Logs.Path != "" {
		logger.Init(agentConfigration.Logs.Path)
	}
	if agentConfigration.Docker.Cloud == "aws" {
		tokens := uitls.GetAwsToken(agentConfigration.Docker.Regin)
		if tokens != nil && len(tokens) == 2 {
			agentConfigration.Docker.Username = tokens[0]
			agentConfigration.Docker.Password = tokens[1]
		}
	}
	StartAgentServer(agentConfigration.Port)
}

// StartAgentServer start the agent
func StartAgentServer(port int16) {
	quit := make(chan int)
	go runRegisterRoutine(quit)
	engine := gin.Default()
	group := engine.Group("/api/v1")
	group.POST("/start/:server", httpStartHandler)
	group.GET("/status/:server", httpStatusHandler)
	group.GET("/status", httpAgentStatus)
	sport := fmt.Sprintf(":%d", port)
	err := engine.Run(sport)
	if err != nil {
		fmt.Println(err)
	}
	quit <- 1
}
