package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

//go:embed assets/*
var assets embed.FS

func main() {
	assetFS, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", HomePage)
	http.HandleFunc("/burn", BurnCPU)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.FS(assetFS)),
		),
	)

	err = http.ListenAndServe(":8443", nil)
	if err != nil {
		panic(err)
	}
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	data, err := assets.ReadFile("assets/index.html")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func BurnCPU(w http.ResponseWriter, r *http.Request) {
	duration := 60
	if s := r.URL.Query().Get("seconds"); s != "" {
		if d, err := strconv.Atoi(s); err == nil && d > 0 {
			duration = d
		}
	}

	workers := runtime.NumCPU()
	if s := r.URL.Query().Get("workers"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			workers = n
		}
	}

	for i := 0; i < workers; i++ {
		go cpuBurn(time.Duration(duration) * time.Second)
	}

	fmt.Fprintf(w, "Started %d CPU workers for %d seconds\n", workers, duration)
}

func cpuBurn(d time.Duration) {
	deadline := time.Now().Add(d)

	var x uint64 = 1
	for time.Now().Before(deadline) {
		// Do some meaningless arithmetic.
		x = x*1664525 + 1013904223

		if x == 0 {
			fmt.Println("Impossible")
		}
	}
}
