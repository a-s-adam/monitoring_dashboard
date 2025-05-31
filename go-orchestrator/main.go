package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	rustMetricsURL   = "http://localhost:8000/metrics"
	pythonAnomalyURL = "http://localhost:8001/detect"
	maxHistorySize   = 20 // Store more history points for better analysis
	port             = "8002"
)

// Rust service response structs
type CpuInfo struct {
	Name      string  `json:"name"`
	Usage     float32 `json:"usage"`
	Frequency uint64  `json:"frequency"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	PercentUsed float32 `json:"percent_used"`
}

type DiskInfo struct {
	Name           string  `json:"name"`
	TotalSpace     uint64  `json:"total_space"`
	AvailableSpace uint64  `json:"available_space"`
	PercentUsed    float32 `json:"percent_used"`
}

type SystemMetrics struct {
	Hostname string     `json:"hostname"`
	Uptime   uint64     `json:"uptime"`
	Cpus     []CpuInfo  `json:"cpus"`
	Memory   MemoryInfo `json:"memory"`
	Disks    []DiskInfo `json:"disks"`
}

// Anomaly detection request/response
type AnomalyRequest struct {
	CpuValues    []float32 `json:"cpu_values"`
	MemoryValues []float32 `json:"memory_values"`
}

type AnomalyResponse struct {
	CpuAnomaly    bool    `json:"cpu_anomaly"`
	MemoryAnomaly bool    `json:"memory_anomaly"`
	CpuScore      float32 `json:"cpu_score,omitempty"`
	MemoryScore   float32 `json:"memory_score,omitempty"`
}

// Combined response for frontend
type DashboardData struct {
	CurrentMetrics SystemMetrics   `json:"current_metrics"`
	Anomalies      AnomalyResponse `json:"anomalies"`
	Timestamp      int64           `json:"timestamp"`
}

// History tracking
type MetricsHistory struct {
	Cpu    []float32
	Memory []float32
	Mutex  sync.Mutex
}

var (
	history      MetricsHistory
	httpClient   = &http.Client{Timeout: 5 * time.Second}
	lastComplete *DashboardData // Cache the last complete response in case a service fails
)

func init() {
	history = MetricsHistory{
		Cpu:    make([]float32, 0, maxHistorySize),
		Memory: make([]float32, 0, maxHistorySize),
	}
}

func fetchHardwareMetrics() (*SystemMetrics, error) {
	resp, err := httpClient.Get(rustMetricsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics from Rust service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rust service returned non-OK status %d: %s",
			resp.StatusCode, string(bodyBytes))
	}

	var metrics SystemMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode metrics response: %w", err)
	}

	// Add to history
	if len(metrics.Cpus) > 0 {
		// Average all CPU usages
		var totalCpuUsage float32
		for _, cpu := range metrics.Cpus {
			totalCpuUsage += cpu.Usage
		}
		avgCpuUsage := totalCpuUsage / float32(len(metrics.Cpus))

		history.Mutex.Lock()
		history.Cpu = append(history.Cpu, avgCpuUsage)
		history.Memory = append(history.Memory, metrics.Memory.PercentUsed)
		if len(history.Cpu) > maxHistorySize {
			history.Cpu = history.Cpu[1:]
			history.Memory = history.Memory[1:]
		}
		history.Mutex.Unlock()
	}

	log.Printf("Fetched metrics - CPU: %.1f%% Memory: %.1f%%",
		metrics.Cpus[0].Usage, metrics.Memory.PercentUsed)
	return &metrics, nil
}

func checkForAnomalies() (*AnomalyResponse, error) {
	history.Mutex.Lock()
	cpuValues := make([]float32, len(history.Cpu))
	memoryValues := make([]float32, len(history.Memory))
	copy(cpuValues, history.Cpu)
	copy(memoryValues, history.Memory)
	history.Mutex.Unlock()

	// If we don't have enough history yet
	if len(cpuValues) < 5 {
		return &AnomalyResponse{
			CpuAnomaly:    false,
			MemoryAnomaly: false,
		}, nil
	}

	reqBody := AnomalyRequest{
		CpuValues:    cpuValues,
		MemoryValues: memoryValues,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal anomaly request: %w", err)
	}

	resp, err := httpClient.Post(pythonAnomalyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to post to Python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("python service returned non-OK status %d: %s",
			resp.StatusCode, string(bodyBytes))
	}

	var anomalyResp AnomalyResponse
	if err := json.NewDecoder(resp.Body).Decode(&anomalyResp); err != nil {
		return nil, fmt.Errorf("failed to decode anomaly response: %w", err)
	}

	if anomalyResp.CpuAnomaly || anomalyResp.MemoryAnomaly {
		log.Printf("Anomaly detected - CPU: %t Memory: %t",
			anomalyResp.CpuAnomaly, anomalyResp.MemoryAnomaly)
	}

	return &anomalyResp, nil
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	metrics, metricsErr := fetchHardwareMetrics()
	anomalies, anomalyErr := checkForAnomalies()

	// Handle cases where one or both services are down
	if metricsErr != nil {
		log.Printf("Error fetching metrics: %v", metricsErr)
		if lastComplete != nil {
			// Return last known good data
			json.NewEncoder(w).Encode(lastComplete)
			return
		}
		http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
		return
	}

	// If anomaly detection fails, we can still return metrics
	if anomalyErr != nil {
		log.Printf("Error checking anomalies: %v", anomalyErr)
		anomalies = &AnomalyResponse{CpuAnomaly: false, MemoryAnomaly: false}
	}

	response := DashboardData{
		CurrentMetrics: *metrics,
		Anomalies:      *anomalies,
		Timestamp:      time.Now().Unix(),
	}

	// Cache the response
	lastComplete = &response

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	history.Mutex.Lock()
	response := map[string]interface{}{
		"cpu":    history.Cpu,
		"memory": history.Memory,
	}
	history.Mutex.Unlock()

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding history: %v", err)
		http.Error(w, "Failed to encode history", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/history", historyHandler)

	log.Printf("Go orchestrator service starting on port :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
