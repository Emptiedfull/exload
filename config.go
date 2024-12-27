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
		Free_ports     []int32 `yaml:"free_ports"`
		Max_load       *int32  `yaml:"max_load"`
		Upscale_ping   *int8   `yaml:"upscale_pings"`
		Downscale_ping *int8   `yaml:"downscale_pings"`
		scale_interval *int    `yaml:"interval"`
	}
	ServerOptions map[string]ServerOption `yaml:"server_options"`
}

type ServerOption struct {
	Prefix          string   `yaml:"prefix"`
	Command         string   `yaml:"command"`
	Args            []string `yaml:"args"`
	Startup_servers *int8    `yaml:"startup_servers"`
	Max_servers     *int8    `yaml:"max_servers"`
}

func getConfig() (Config, error) {
	var default_port int32 = 8080
	default_static := "/static"
	var default_max_servers int8 = int8(127)
	var default_Max_load int32 = 100
	var default_start_servers int8 = 2
	var config Config
	var default_pings int8 = 2
	var default_interval int = 3

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

	for key, srv := range config.ServerOptions {
		if srv.Startup_servers == nil {
			srv.Startup_servers = &default_start_servers
		}

		if srv.Max_servers == nil {
			srv.Max_servers = &default_max_servers
		}

		config.ServerOptions[key] = srv
	}

	if config.Proxy_settings.Downscale_ping == nil {
		config.Proxy_settings.Downscale_ping = &default_pings
	}
	if config.Proxy_settings.Upscale_ping == nil {
		config.Proxy_settings.Upscale_ping = &default_pings
	}

	if config.Proxy_settings.scale_interval == nil {
		config.Proxy_settings.scale_interval = &default_interval
	}

	fmt.Println(*config.Proxy_settings.scale_interval)

	return config, nil
}
