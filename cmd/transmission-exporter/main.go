package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/metalmatze/transmission-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	addr := env("TRANSMISSION_ADDR", "http://localhost:9091")
	username := os.Getenv("TRANSMISSION_USERNAME")
	password := os.Getenv("TRANSMISSION_PASSWORD")
	webAddr := env("WEB_ADDR", ":19091")
	webPath := env("WEB_PATH", "/metrics")

	slog.Info("starting transmission-exporter",
		"transmission_addr", addr,
		"web_addr", webAddr,
		"web_path", webPath,
		"auth_enabled", username != "",
	)

	var user *transmission.User
	if username != "" && password != "" {
		user = &transmission.User{
			Username: username,
			Password: password,
		}
	}

	client := transmission.New(addr, user)

	prometheus.MustRegister(NewTorrentCollector(client))
	prometheus.MustRegister(NewSessionCollector(client))
	prometheus.MustRegister(NewSessionStatsCollector(client))

	http.Handle(webPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Transmission Exporter</title></head>
			<body>
			<h1>Transmission Exporter</h1>
			<p><a href="` + webPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	slog.Info("listening", "addr", webAddr)
	if err := http.ListenAndServe(webAddr, nil); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func boolToString(true bool) string {
	if true {
		return "1"
	}
	return "0"
}
