package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type GRPC struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Log struct {
	Level string `yaml:"level"`
}

type Config struct {
	GRPC GRPC `yaml:"grpc"`
	Log  Log  `yaml:"log"`
}

func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
