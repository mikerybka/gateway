package main

import (
	"encoding/json"
	"os"
	"strings"
)

type Config map[string]Backend
type Backend struct {
	Port string `json:"port"`
	Auth struct {
		User string `json:"user"`
		Pass string `json:"pass"`
		Dir  string `json:"dir"`
	} `json:"auth"`
}

func main() {
	cfgPath := os.Args[1]
	oldConfig := map[string]string{}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &oldConfig)
	if err != nil {
		panic(err)
	}
	newConfig := Config{}
	for k, v := range oldConfig {
		newConfig[k] = Backend{
			Port: strings.TrimPrefix(v, "http://localhost:"),
		}
	}
	b, err = json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(cfgPath, b, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
