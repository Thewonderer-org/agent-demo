package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var raceStrategyServiceURL = "http://localhost:7004"

func init() {
	if u := os.Getenv("RACE_STRATEGY_SERVICE_URL"); u != "" {
		raceStrategyServiceURL = u
	}
}

type PitstopPlan struct {
	Driver    string `json:"driver"`
	Team      string `json:"team"`
	PlannedLap int   `json:"planned_pit_lap"`
	TireChoice string `json:"tire_choice"`
	Priority   string `json:"priority"`
}

type OrchestratorReport struct {
	GeneratedAt string      `json:"generated_at"`
	Strategies  interface{} `json:"strategies"`
	PitstopPlan []PitstopPlan `json:"pitstop_plan"`
}

var (
	cachedReport *OrchestratorReport
	cacheMu      sync.RWMutex
)

func fetchStrategies() (interface{}, error) {
	resp, err := http.Get(raceStrategyServiceURL + "/api/strategies")
	if err != nil {
		return nil, fmt.Errorf("failed to reach race-strategy-service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	json.Unmarshal(body, &result)
	return result, nil
}

func buildPitstopPlan() []PitstopPlan {
	return []PitstopPlan{
		{Driver: "Max Verstappen", Team: "Red Bull", PlannedLap: 24, TireChoice: "Hard", Priority: "high"},
		{Driver: "Lewis Hamilton", Team: "Mercedes", PlannedLap: 28, TireChoice: "Medium", Priority: "medium"},
		{Driver: "Charles Leclerc", Team: "Ferrari", PlannedLap: 17, TireChoice: "Medium", Priority: "high"},
		{Driver: "Lando Norris", Team: "McLaren", PlannedLap: 23, TireChoice: "Hard", Priority: "medium"},
	}
}

func pollRaceStrategyService() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	refresh := func() {
		strategies, err := fetchStrategies()
		if err != nil {
			log.Printf("[poll] could not fetch strategies: %v", err)
		}

		report := &OrchestratorReport{
			GeneratedAt: time.Now().Format(time.RFC3339),
			Strategies:  strategies,
			PitstopPlan: buildPitstopPlan(),
		}

		cacheMu.Lock()
		cachedReport = report
		cacheMu.Unlock()
		log.Println("[poll] refreshed data from race-strategy-service")
	}

	refresh()
	for range ticker.C {
		refresh()
	}
}

func handleOrchestration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cacheMu.RLock()
	report := cachedReport
	cacheMu.RUnlock()

	if report == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "data not yet available"})
		return
	}

	json.NewEncoder(w).Encode(report)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "pitstop-orchestrator-service",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func main() {
	go pollRaceStrategyService()

	http.HandleFunc("/api/orchestration", handleOrchestration)
	http.HandleFunc("/health", handleHealth)

	log.Println("pitstop-orchestrator-service starting on :7005")
	log.Fatal(http.ListenAndServe(":7005", nil))
}
