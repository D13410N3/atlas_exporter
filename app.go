package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

type AtlasProbeInfo struct {
	Status       map[string]interface{} `json:"status"`
	StatusSince  float64                `json:"status_since"`
	TotalUptime  float64                `json:"total_uptime"`
	AddressV4    string                 `json:"address_v4"`
	AddressV6    string                 `json:"address_v6"`
	FirstConnected float64              `json:"first_connected"`
	LastConnected  float64              `json:"last_connected"`
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("Getting info for atlas id %s", id)

	resp, err := http.Get(fmt.Sprintf("https://atlas.ripe.net/api/v2/probes/%s/", id))
	if err != nil {
		log.Println("Error while fetching data:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer resp.Body.Close()

	var probeInfo AtlasProbeInfo
	if err := json.NewDecoder(resp.Body).Decode(&probeInfo); err != nil {
		log.Println("Error while decoding JSON:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	statusMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_status",
		Help: "Atlas Probe Status ID",
	})
	statusMetric.Set(probeInfo.Status["id"].(float64))

	statusSinceMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_status_since",
		Help: "Atlas Probe Status Since",
	})
	statusSinceMetric.Set(probeInfo.StatusSince)

	totalUptimeMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_total_uptime",
		Help: "Atlas Probe Total Uptime",
	})
	totalUptimeMetric.Set(probeInfo.TotalUptime)

	firstConnectedMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_first_connected",
		Help: "Atlas Probe First Connected",
	})
	firstConnectedMetric.Set(probeInfo.FirstConnected)

	lastConnectedMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_last_connected",
		Help: "Atlas Probe Last Connected",
	})
	lastConnectedMetric.Set(probeInfo.LastConnected)

	infoMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "atlas_info",
		Help: "Atlas Probe Info",
	}, []string{"address_v4", "address_v6"})

	infoMetric.WithLabelValues(probeInfo.AddressV4, probeInfo.AddressV6).Set(1)

	registry := prometheus.NewRegistry()
	registry.MustRegister(statusMetric, statusSinceMetric, totalUptimeMetric, firstConnectedMetric, lastConnectedMetric, infoMetric)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "0.0.0.0:9207"
	}

	r := mux.NewRouter()
	r.HandleFunc("/metrics/{id:[0-9]+}", metricsHandler)
	http.Handle("/", r)

	log.Printf("Starting server on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
