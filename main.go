package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func main() {

	config, err := getConfig()
	if err != nil {
		fmt.Println("Error loading config")
		os.Exit(1)
	}

	m := NewManager(config)
	defer m.file.Close()

	fs := http.FileServer(http.Dir(*config.Static_path))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println(m.servers)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.ServeFile(w, r, "static/favicon.ico")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		m.proxy(url, w, r)
	})

	http.ListenAndServe(":"+strconv.Itoa(int(*config.Proxy_port)), nil)
}
