package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"cloudeng.io/go/cmd/goannotate/annotators"
	"gopkg.in/yaml.v2"
)

/*
type AnnotationConfig struct {
	Name                 string `yaml:"name"`
	ContextType          string `yaml:"contextType"`
	Function             string `yaml:"function"`
	Import               string `yaml:"import"`
	Tag                  string `yaml:"tag"`
	IgnoreEmptyFunctions bool   `yaml:"ignoreEmptyFunctions"`
}

	Interfaces  []string     `yam:"interfaces"`
	Functions   []string     `yam:"functions"`

type Annotation struct {
	Name string `yaml:"name"`
	yaml.MapSlice
}
*/

type Debug struct {
	CPUProfile string `yaml:"cpu_profile"`
}

//type Options struct {
//	Concurrency int `yaml:"concurrency"`
//}

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
	/*	if err := json.Unmarshal(buf, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %v: %v", filename, err)
	}*/
	///	config.configureDefaults()
	return config, err
}

/*func (c *Config) configureDefaults() {
	if c.Options.Concurrency == 0 {
		c.Options.Concurrency = runtime.NumCPU()
	}
}*/

/*
func (c *Config) String() string {
	out := &strings.Builder{}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.Encode(c)
	return out.String()
}*/
