package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AtlasProbe struct {
	AddressV4    string `json:"address_v4"`
	AddressV6    string `json:"address_v6"`
	Status       Status `json:"status"`
	StatusSince  int64  `json:"status_since"`
	TotalUptime  int64  `json:"total_uptime"`
}

type Status struct {
	ID int `json:"id"`
}

func fetchMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Getting info for atlas id %s", id)

	resp, err := http.Get(fmt.Sprintf("https://atlas.ripe.net/api/v2/probes/%s/", id))
	if err != nil {
		log.Printf("Failed to fetch data: %s", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read data: %s", err)
		http.Error(w, "Failed to read data", http.StatusInternalServerError)
		return
	}

	var probe AtlasProbe
	if err := json.Unmarshal(body, &probe); err != nil {
		log.Printf("Failed to parse data: %s", err)
		http.Error(w, "Failed to parse data", http.StatusInternalServerError)
		return
	}

	// Create a new registry for this request
	registry := prometheus.NewRegistry()

	atlasStatus := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_status",
		Help: "Atlas Probe Status ID",
	})
	registry.MustRegister(atlasStatus)

	atlasStatusSince := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_status_since",
		Help: "Atlas Probe Status Since",
	})
	registry.MustRegister(atlasStatusSince)

	atlasTotalUptime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "atlas_total_uptime",
		Help: "Atlas Probe Total Uptime",
	})
	registry.MustRegister(atlasTotalUptime)

	atlasInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "atlas_info",
		Help: "Atlas Probe Info",
	}, []string{"address_v4", "address_v6"})
	registry.MustRegister(atlasInfo)

	// Set the values for the metrics in this request
	atlasStatus.Set(float64(probe.Status.ID))
	atlasStatusSince.Set(float64(probe.StatusSince))
	atlasTotalUptime.Set(float64(probe.TotalUptime))
	atlasInfo.WithLabelValues(probe.AddressV4, probe.AddressV6).Set(1)

	// Serve metrics from the local registry
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "0.0.0.0:9207"
	}

	r := mux.NewRouter()
	r.HandleFunc("/metrics/{id:[0-9]+}", fetchMetrics)
	http.Handle("/", r)

	log.Printf("Server is running on %s", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}
