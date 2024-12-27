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

		m.proxy(url, w, r)
	})

	pid := os.Getpid()
	fmt.Printf("The PID of this process is %d\n", pid)

	http.ListenAndServe(":"+strconv.Itoa(int(*config.Proxy_port)), nil)
}
