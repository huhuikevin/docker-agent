package uitls

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
	"jwaoo.com/logger"
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
