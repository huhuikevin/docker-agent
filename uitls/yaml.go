package uitls

import (
	"io/ioutil"

	"github.com/huhuikevin/docker-agent/logger"
	"gopkg.in/yaml.v2"
)

//LoadYAML load yaml tools
func LoadYAML(filename string, v interface{}) error {
	logger.Println("configfile", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Println(err)
		return err
	}
	err = yaml.Unmarshal(data, v)
	if err != nil {
		logger.Println(err)
		return err
	}
	return nil
}
