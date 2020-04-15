// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

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
	Annotations []annotators.Spec `yam:"annotations"`
	Debug       Debug             `yam:"debug"`
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
