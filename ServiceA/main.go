package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type cepRequest struct {
	Cep string `json:"cep"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for:", r.URL.Path)
	var req cepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	cep := req.Cep

	if cep == "" {
		log.Println("No CEP provided in the request")
		http.Error(w, "CEP is required", http.StatusBadRequest)
		return
	}
	log.Println("Extracted CEP:", cep)
	temperature, err := getTemperature(cep)
	if err != nil {
		log.Printf("Error getting temperature: %v\n", err)
		http.Error(w, "Failed to get temperature", http.StatusInternalServerError)
		return
	}

	log.Println("Temperature data received:", temperature)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(temperature))
	if err != nil {
		log.Printf("Error writing response: %v\n", err)
		return
	}
	log.Println("Response sent successfully")
}

func getTemperature(cep string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://127.0.0.1:8081/temperature/%s", cep), nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for service B: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error during request to service B: %w", err)
	}

	defer resp.Body.Close()

	log.Println("Response from service B:", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body response: %w", err)
	}
	log.Println("Response body from service B:", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("service B returned status code %d with body: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/temperature", handler)

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("An error occurred while starting the server: %v", err)
	}
}
