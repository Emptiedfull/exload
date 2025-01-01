package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

var con Config

type Config struct {
	Admin_port  *int    `yaml:"admin_port`
	Proxy_port  *int32  `yaml:"proxy_port"`
	Static_path *string `yaml:"static_path"`
	Dynos       struct {
		Scaler  bool `yaml:"scaler"`
		Monitor bool `yaml:"monitor"`
	}

	Scaling_settings struct {
		Max_load       *int32 `yaml:"max_load"`
		Min_Load       *int   `yaml:"min_load"`
		Upscale_ping   *int   `yaml:"upscale_pings"`
		Downscale_ping *int   `yaml:"downscale_pings"`
		scale_interval *int   `yaml:"interval"`
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

func getConfig() error {
	var default_admin_port int = 9000
	var default_port int32 = 8080
	default_static := "/static"
	var default_max_servers int8 = int8(127)
	var default_Max_load int32 = 100
	var default_Min_load int = 90
	var default_start_servers int8 = 2
	var config Config
	var default_pings int = 2
	var default_interval int = 3

	file, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("huh")
		return err
	}

	if config.Proxy_port == nil {
		config.Proxy_port = &default_port
	}

	if config.Admin_port == nil {
		config.Admin_port = &default_admin_port
	}

	if config.Static_path == nil {
		config.Static_path = &default_static
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

	if config.Scaling_settings.Max_load == nil {
		config.Scaling_settings.Max_load = &default_Max_load
	}

	fmt.Println("pings", config.Scaling_settings.Upscale_ping)

	if config.Scaling_settings.Downscale_ping == nil {
		config.Scaling_settings.Downscale_ping = &default_pings
	}
	if config.Scaling_settings.Upscale_ping == nil {
		config.Scaling_settings.Upscale_ping = &default_pings
	}

	if config.Scaling_settings.scale_interval == nil {
		config.Scaling_settings.scale_interval = &default_interval
	}

	if config.Scaling_settings.Min_Load == nil {
		config.Scaling_settings.Min_Load = &default_Min_load
	}

	con = config

	return nil
}
