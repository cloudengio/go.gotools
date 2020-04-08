package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"cloudeng.io/go/cmd/goannotate/annotators"
	"gopkg.in/yaml.v2"
)

type Debug struct {
	CPUProfile string `yaml:"cpu_profile"`
}

type Config struct {
	Packages   []string          `yam:"packages"`
	Annotators []annotators.Spec `yam:"annotators"`
	Debug      Debug             `yam:"debug"`
}

func ConfigFromFile(filename string) (*Config, error) {
	config := &Config{}
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return config, err
}
