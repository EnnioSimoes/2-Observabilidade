package main

import (
	"log"
	"net/http"

	"encoding/json"

	address "github.com/EnnioSimoes/2-Observabilidade/ServiceB/address"
	weather "github.com/EnnioSimoes/2-Observabilidade/ServiceB/weather"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func handler(w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")
	if cep == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	addr, err := address.GetCep(cep)
	if err != nil {
		log.Println("Error getting address:", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	if addr.Cep == "" {
		log.Println("Error: Address not found for zipcode:", cep)
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "can not find zipcode"})
		return
	}

	temperature, err := weather.GetWeather(addr.Localidade)
	if err != nil {
		log.Println("Error getting temperature:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(temperature)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/temperature/{cep}", handler)

	log.Println("Starting server on :8081")
	err := http.ListenAndServe(":8081", r)
	if err != nil {
		log.Fatalf("An error occurred while starting the server: %v", err)
	}
}
