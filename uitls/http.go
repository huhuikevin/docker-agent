package uitls

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"jwaoo.com/logger"
)

//ResultData generic result data
type ResultData struct {
	StatusCode int
	Data       []byte
}

//DoHTTPGet http get
func DoHTTPGet(url string) (*ResultData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err.Error())
		return nil, err
	}
	return &ResultData{StatusCode: resp.StatusCode, Data: respBytes}, nil
}

//DoHTTPPost http post
func DoHTTPPost(url string, params interface{}) (*ResultData, error) {
	jsonStu, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(jsonStu)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Println(err.Error())
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err.Error())
		return nil, err
	}
	return &ResultData{StatusCode: resp.StatusCode, Data: respBytes}, nil
}
