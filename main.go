package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

func main() {

	adminServer := http.NewServeMux()

	logs, _ := os.ReadDir("./logs/server_logs")
	for _, file := range logs {
		path := "./logs/server_logs/" + file.Name()
		os.RemoveAll(path)
	}

	err := getConfig()

	if err != nil {
		fmt.Println("Error loading config")
		os.Exit(1)
	}
	c := NewCache(1)
	// c.put("/api/ping", make([]byte, 5), 1*time.Hour)
	m := NewManager()
	defer m.file.Close()

	fs := http.FileServer(http.Dir("./static"))
	adminServer.Handle("/static/", http.StripPrefix("/static", fs))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.ServeFile(w, r, "static/favicon.ico")
	})

	adminServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

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

	port := strconv.Itoa(int(*con.Proxy_port))
	fmt.Printf("Starting server on port %s\n", port)

	go func() {
		err = http.ListenAndServe(":"+strconv.Itoa(*con.Admin_port), adminServer)
	}()

	if con.Statics.Fileserver != nil {
		fs := http.FileServer(http.Dir(*con.Statics.Fileserver))
		http.Handle("/static/", http.StripPrefix("/static", fs))
	}

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
