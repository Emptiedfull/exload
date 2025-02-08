package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func main() {

	adminServer := http.NewServeMux()

	err := getConfig()
	conns := make(map[string][]*client)

	if err != nil {
		fmt.Println("Error loading config")
		os.Exit(1)

		// c.put("/api/ping", make([]byte, 5), 1*time.Hour)

	}
	fs := http.FileServer(http.Dir("./static"))
	adminServer.Handle("/static/", http.StripPrefix("/static", fs))
	go monitorDyno(conns, make(<-chan bool))
	adminServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("handling", r.URL.String())
		monitor(conns, r.URL.String(), w, r)

	})
	fmt.Println(strconv.Itoa(*con.Admin_port))

	err = http.ListenAndServe(":"+strconv.Itoa(*con.Admin_port), adminServer)

}
