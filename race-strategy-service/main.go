package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Strategy struct {
	Driver      string `json:"driver"`
	Team        string `json:"team"`
	TireCompound string `json:"tire_compound"`
	Stint       int    `json:"stint_laps"`
	PitWindow   string `json:"pit_window"`
}

var strategies = []Strategy{
	{Driver: "Max Verstappen", Team: "Red Bull", TireCompound: "Medium", Stint: 25, PitWindow: "Lap 20-28"},
	{Driver: "Lewis Hamilton", Team: "Mercedes", TireCompound: "Hard", Stint: 30, PitWindow: "Lap 25-32"},
	{Driver: "Charles Leclerc", Team: "Ferrari", TireCompound: "Soft", Stint: 18, PitWindow: "Lap 15-20"},
	{Driver: "Lando Norris", Team: "McLaren", TireCompound: "Medium", Stint: 24, PitWindow: "Lap 20-26"},
}

func handleStrategies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "race-strategy-service",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func main() {
	http.HandleFunc("/api/strategies", handleStrategies)
	http.HandleFunc("/health", handleHealth)

	log.Println("race-strategy-service starting on :7004")
	log.Fatal(http.ListenAndServe(":7004", nil))
}
