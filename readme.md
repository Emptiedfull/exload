# Exload

Exload is a lightweight reverse proxy made in go with extensive support and tooling.

## Features

- multipathing support
- health checks and fault tolerance
- Real-time monitoring of server metrics
- Dynamic scaling of servers
- Unix socket support for internal servers
- support for static and external proxying too
- WebSocket support
- Interactive charts for visualizing server performance
- In memory caching with custom cache control headers
- Configurable settings via `config.yaml`

## Getting Started

### Prerequisities

- Go 1.23.4 or later
- Python 3.x (optional, only required for demo)

### Installation 

1. Clone the repository:    
     ```sh
    git clone https://github.com/yourusername/exload.git
    cd exload
    ```

2. Install Go dependecies
    ```sh
    go mod download
    ```

### Running the proxy

1. Configure config.yaml 

    ``` yaml
    # Port of proxy access point
    proxy_port: 80
    # Port for statistics and admin pane; 
    admin_port: 9000
    #optional addons
    dynos: 
        scaler: true #enables dynamic scaling
        monitor: true #enables monitoring
  
    scaling_settings: #only applies when scaler dyno is active

        #Load = requests per second/no. of servers

        #pings upscale after this threshold
        max_load: 50 
        #pings downscale after this threshold
        min_load: 10
        #pings required to start upscale
        upscale_pings: 5
        #pings required to start downscale
        downscale_pings : 600
        #interval for checking load
        interval : 1

    server_options: #servers started by the proxy it self
        server2:
            prefix: /api #server endpoint
            command: venv/bin/uvicorn #root command to be executed
            args: ["server:app","--uds","<sock>"] #arguements, must contain <sock>
            startup_servers: 4 #servers at startup
            max_servers: 6 #maximum servers
        server1: #another server to be added
            prefix: /admin
            command: venv/bin/uvicorn
            args: ["server:app","--uds","<sock>"]
            startup_servers: 1
            max_servers: 4

    statics: #static options
        fileserver: "static" #centralized static path
        servers: #servers not started by proxy
            joker:
                type: "port" #internal
                access: "8001" #access port
                prefix : '/joker' #endpoint
            koker:  
                type: "external" #external
                access: "http://google.com" #access url
                prefix: "/koker" #endpoints
    ```

2. Start the server 
    ```sh
    go run main.go
    ```

All requests at :80 automatically get forwarded to the configured endpoints

### Testing options

1. Go Test Suite: Allows for more load at one point
2. Python stress.py: Lower system resource usage, lesser load 

### acknowledgement 

This project has used github copilot as a debugging tool