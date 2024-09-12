package config

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	ListenPort     int    `json:"listenPort"`
	ListenAddress  string `json:"listenAddress"`
	CPUSensorChip  string `json:"cpuSensorChip"`
	Manual         bool   `json:"manual"`
	Frontend       bool   `json:"frontend"`
	RefreshOnStart bool   `json:"refreshOnStart"`
	Metrics        bool   `json:"metrics"`
	DbusMonitor    bool   `json:"dbusMonitor"`
}

var configuration Configuration

// Init will initialize a new config object
func Init() {
	pwd, _ := os.Getwd()
	cfg := pwd + "/config.json"
	f, err := os.Open(cfg)
	if err != nil {
		panic(err.Error())
	}
	if err = json.NewDecoder(f).Decode(&configuration); err != nil {
		panic(err.Error())
	}
}

// GetConfig will return structs.Configuration struct
func GetConfig() Configuration {
	return configuration
}
