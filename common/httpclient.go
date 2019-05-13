package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	"jwaoo.com/uitls"
	//"unsafe"
)

func httpResultData(result *uitls.ResultData) *RequstResult {
	if result.StatusCode == http.StatusOK || result.StatusCode == http.StatusBadRequest {
		respBytes := result.Data
		//byte数组直接转成string，优化内存
		//str := (*string)(unsafe.Pointer(&respBytes))
		//fmt.Println(*str)
		resultData := RequstResult{}
		if err := json.Unmarshal(respBytes, &resultData); err != nil {
			fmt.Println(err)
			errResult := NewResult(CannotMapToResult)
			errResult.Data = err.Error()
			return errResult
		}
		return &resultData
	}
	msg := fmt.Sprintf("Http Error code = %d", result.StatusCode)
	errResult := NewResult(UnKnowServerError)
	errResult.Data = msg
	return errResult
}

//DoHTTPPostJSON post json
func DoHTTPPostJSON(url string, data interface{}) *RequstResult {
	result, err := uitls.DoHTTPPost(url, data)
	if err != nil {
		fmt.Println(err.Error())
		result := NewResult(LocalNetworkError)
		return result
	}

	return httpResultData(result)
}

//DoHTTPGet http get
func DoHTTPGet(url string) *RequstResult {
	result, err := uitls.DoHTTPGet(url)
	if err != nil {
		result := NewResult(LocalNetworkError)
		return result
	}
	return httpResultData(result)
}
