package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goinggo/mapstructure"

	"github.com/huhuikevin/docker-agent/common"

	"github.com/huhuikevin/docker-agent/clients"
)

func main() {
	pserver := flag.String("server", "", "server to start")
	ptag := flag.String("tag", "", "docker image's tag")
	pconfigdir := flag.String("path", "config", "path for config")
	flag.Parse()

	configdir := *pconfigdir
	tag := *ptag
	server := *pserver
	parts := strings.Split(server, ":")
	if len(parts) == 2 {
		server = parts[0]
		tag = parts[1]
	}
	if server == "" && tag == "" {
		fmt.Println("param error!!!")
		os.Exit(1)
	}
	//appclient.LoadConfigfile("config/runjava.json")
	config := clients.LoadConfigfileByServer(configdir, server)
	params := clients.BuildParamsByServerName(*config, server, tag)
	url := fmt.Sprintf("%s/api/v1/start/%s", config.ProxyHTTP, server)
	for _, param := range params {
		fmt.Println("params =", param)
		err := startDocker(url, param)
		if err != nil {
			//fmt.Println("err = ", err.Error())
			os.Exit(1)
		}
		err = checkHealth(server, param)
		if err != nil {
			fmt.Println("*FAILT ==> ", err.Error())
			os.Exit(1)
		}
	}
}

func checkHealth(server string, info map[string]interface{}) error {
	_, ok := info["checkHealth"]
	if !ok {
		fmt.Println("Not checkHealth")
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/check/%s", clients.GetProxyHTTPURL(), server)
	count := 30
	for {
		result := common.DoHTTPPostJSON(url, info)
		if result.Code == common.Success {
			fmt.Println("SUCCESS ==> ", result.Message)
			return nil
		}
		fmt.Println("Wait for Healthy ............")
		data := result.Data.([]common.RequstResult)
		for _, d := range data {
			fmt.Println(d.Message)
		}
		count = count - 1
		if count <= 0 {
			return errors.New("check health error")
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func startDocker(url string, info map[string]interface{}) error {
	count := 5
	for {
		result := common.DoHTTPPostJSON(url, info)
		if result.Code != common.Success {
			fmt.Println(result.Message, "Try again")
		} else {
			fmt.Println("SUCCESS ==> ", result.Message)
			return nil
		}

		details := getResultsFromResultData(result.Data)
		fmt.Println("Details reason is:")
		for _, detail := range details {
			fmt.Println("on host", detail.Host, detail.Message, ":", detail.Data)
		}
		count = count - 1
		if count <= 0 {
			return errors.New("start docker err")
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func getResultsFromResultData(resultdata interface{}) []common.RequstResult {
	result := make([]common.RequstResult, 0)
	data := resultdata.([]interface{})
	for _, d := range data {
		dict := d.(map[string]interface{})
		var detail common.RequstResult
		err := mapstructure.Decode(dict, &detail)
		if err != nil {
			fmt.Println("decode error=", err)
		} else {
			result = append(result, detail)
		}
	}
	return result
}
