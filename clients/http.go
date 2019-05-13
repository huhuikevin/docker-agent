package clients

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/huhuikevin/docker-agent/uitls"

	//"unsafe"
	"errors"
)

//DoHTTPPostJSON post json
func DoHTTPPostJSON(url string, data map[string]interface{}) (string, error) {
	result, err := uitls.DoHTTPPost(url, data)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	if result.StatusCode == http.StatusOK || result.StatusCode == http.StatusInternalServerError {
		respBytes := result.Data
		//byte数组直接转成string，优化内存
		//str := (*string)(unsafe.Pointer(&respBytes))
		//fmt.Println(*str)
		jsonData := make(map[string]string)
		if err := json.Unmarshal(respBytes, &jsonData); err == nil {
			if result.StatusCode == http.StatusOK {
				return jsonData["status"], nil
			}
			return "", errors.New(jsonData["status"])
		}
		fmt.Println(err)
		return "", err
	}
	msg := fmt.Sprintf("Http Error code = %d", result.StatusCode)
	return "", errors.New(msg)
}

//DoHTTPGet http get
func DoHTTPGet(url string) (string, error) {
	result, err := uitls.DoHTTPGet(url)
	if err != nil {
		return "", err
	}
	if result.StatusCode == http.StatusOK || result.StatusCode == http.StatusInternalServerError {
		respBytes := result.Data
		if err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		//byte数组直接转成string，优化内存
		//str := (*string)(unsafe.Pointer(&respBytes))
		//fmt.Println(*str)
		jsonData := make(map[string]string)
		if err := json.Unmarshal(respBytes, &jsonData); err == nil {
			if result.StatusCode == http.StatusOK {
				return jsonData["status"], nil
			}
			return "", errors.New(jsonData["status"])
		}
		fmt.Println(err)
		return "", err
	}
	msg := fmt.Sprintf("Http Error code = %d", result.StatusCode)
	return "", errors.New(msg)
}
