package parser

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Settings struct {
	Databases []struct {
		Alias    string `yaml:"alias"`
		Type     string `yaml:"type"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"databases"`
}

type Seed struct {
	DB   string `yaml:"db"`
	File string `yaml:"file"`
}

type Hook struct {
	Target   string            `yaml:"target"`
	Headers  map[string]string `yaml:"headers"`
	Captures map[string]string `yaml:"captures"`
}

type BeforeEach struct {
	Seeds []Seed `yaml:"seeds"`
	Hooks []Hook `yaml:"hooks"`
}

type Expect struct {
	Status int            `yaml:"status"`
	Body   map[string]any `yaml:"body"`
}

type DBCheck struct {
	DB     string           `yaml:"db"`
	Query  string           `yaml:"query"`
	Expect []map[string]any `yaml:"expect"`
}

type Test struct {
	Name     string            `yaml:"name"`
	Target   string            `yaml:"target"`
	Headers  map[string]string `yaml:"headers"`
	Expect   Expect            `yaml:"expect"`
	DBChecks []DBCheck         `yaml:"dbChecks"`
}

type Suite struct {
	Settings   Settings   `yaml:"settings"`
	BeforeEach BeforeEach `yaml:"beforeEach"`
	Tests      []Test     `yaml:"tests"`
}

func ParseSuite(file string) (*Suite, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var config Suite
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
