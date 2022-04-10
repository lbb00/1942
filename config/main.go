package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	Cookie   string `yaml:"cookie"`
	Uid      string `yaml:"uid"`
	DeviceId string `yaml:"deviceId"`

	Longitude string `yaml:"longitude"`
	Latitude  string `yaml:"latitude"`

	BarkId string `yaml:"barkId"`
}

func getYamlPath(yamlName string) string {
	str, _ := os.Getwd()

	println(str)

	return str + "/" + yamlName + ".yml"
}

func (c *Conf) GetConf(yamlName string) (*Conf, error) {
	yamlPath := getYamlPath(yamlName)

	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(yamlFile, c)

	if err != nil {
		return c, err
	}

	return c, nil
}
