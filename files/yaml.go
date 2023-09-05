package yaml

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadYAML(file string, cnf interface{}) error {
	yamlFile, err := os.ReadFile(file)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, cnf)
	}
	return err
}
