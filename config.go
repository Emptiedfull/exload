package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Proxy_port     *int32  `yaml:"port"`
	Static_path    *string `yaml:"static_path"`
	Proxy_settings struct {
		Free_ports []int32 `yaml:"free_ports"`
		Max_load   *int32  `yaml:"max_load"`

		Scale_pings *int `yaml:"scale_pings"`
	}
	ServerOptions map[string]ServerOption `yaml:"server_options"`
}

type ServerOption struct {
	Prefix          string   `yaml:"prefix"`
	Command         string   `yaml:"command"`
	Args            []string `yaml:"args"`
	Startup_servers *int8    `yaml:"startup_servers"`
}

func getConfig() (Config, error) {
	var default_port int32 = 8080
	default_static := "/static"
	var default_Max_load int32 = 100
	var default_start_servers int8 = 2
	var config Config
	var default_pings = 2

	file, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return Config{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return Config{}, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("huh")
		return Config{}, err
	}

	if config.Proxy_port == nil {
		config.Proxy_port = &default_port
	}

	if config.Static_path == nil {
		config.Static_path = &default_static
	}

	if config.Proxy_settings.Max_load == nil {
		config.Proxy_settings.Max_load = &default_Max_load
	}

	for _, srv := range config.ServerOptions {
		if srv.Startup_servers == nil {
			srv.Startup_servers = &default_start_servers
		}
	}

	if config.Proxy_settings.Scale_pings == nil {
		config.Proxy_settings.Scale_pings = &default_pings
	}

	return config, nil

}
