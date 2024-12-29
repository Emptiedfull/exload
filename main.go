package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

func main() {

	logs, _ := os.ReadDir("./logs/server_logs")
	for _, file := range logs {
		path := "./logs/server_logs/" + file.Name()
		os.RemoveAll(path)
	}

	config, err := getConfig()
	if err != nil {
		fmt.Println("Error loading config")
		os.Exit(1)
	}
	c := NewCache(100)
	// c.put("/api/ping", make([]byte, 5), 1*time.Hour)
	m := NewManager(config)
	defer m.file.Close()

	fmt.Println("manager loaded")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.ServeFile(w, r, "static/favicon.ico")
	})

	http.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {

		monitor(m, r.URL.String(), w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		if websocket.IsWebSocketUpgrade(r) {
			m.proxyWebSocket(w, r, url)
		} else {
			m.proxy(url, w, r, c)
		}

	})

	pid := os.Getpid()
	fmt.Printf("The PID of this process is %d\n", pid)

	port := strconv.Itoa(int(*config.Proxy_port))
	fmt.Printf("Starting server on port %s\n", port)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
