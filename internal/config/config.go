package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	HTTP struct {
		Port string `yaml:"port"`
	} `yaml:"http"`
	DB struct {
		Port     string `yaml:"port"`
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"db"`
}

func LoadYamlConfig() *Config {
	env := "development"
	if e := os.Getenv("ENV"); e != "" {
		env = e
	}

	yamlFile, err := ioutil.ReadFile(fmt.Sprintf("./config/%s.yaml", env))
	if err != nil {
		log.Printf("yamlFile get err: %v\n", err)
		os.Exit(1)
	}

	cfg := Config{}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Printf("Unmarshal yamlFile err: %v\n", err)
		os.Exit(1)
	}

	return &cfg
}
