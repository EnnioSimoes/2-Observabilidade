package weather

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/EnnioSimoes/2-Observabilidade/ServiceB/configs"
)

type Weatherapi struct {
	Location struct {
		Name           string  `json:"name"`
		Region         string  `json:"region"`
		Country        string  `json:"country"`
		Lat            float64 `json:"lat"`
		Lon            float64 `json:"lon"`
		TzID           string  `json:"tz_id"`
		LocaltimeEpoch int     `json:"localtime_epoch"`
		Localtime      string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int     `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            float64 `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMph    float64 `json:"wind_mph"`
		WindKph    float64 `json:"wind_kph"`
		WindDegree int     `json:"wind_degree"`
		WindDir    string  `json:"wind_dir"`
		PressureMb float64 `json:"pressure_mb"`
		PressureIn float64 `json:"pressure_in"`
		PrecipMm   float64 `json:"precip_mm"`
		PrecipIn   float64 `json:"precip_in"`
		Humidity   int     `json:"humidity"`
		Cloud      int     `json:"cloud"`
		FeelslikeC float64 `json:"feelslike_c"`
		FeelslikeF float64 `json:"feelslike_f"`
		WindchillC float64 `json:"windchill_c"`
		WindchillF float64 `json:"windchill_f"`
		HeatindexC float64 `json:"heatindex_c"`
		HeatindexF float64 `json:"heatindex_f"`
		DewpointC  float64 `json:"dewpoint_c"`
		DewpointF  float64 `json:"dewpoint_f"`
		VisKm      float64 `json:"vis_km"`
		VisMiles   float64 `json:"vis_miles"`
		Uv         float64 `json:"uv"`
		GustMph    float64 `json:"gust_mph"`
		GustKph    float64 `json:"gust_kph"`
	} `json:"current"`
}

type Temperature struct {
	Temp_C float64 `json:"temp_c"`
	Temp_K float64 `json:"temp_k"`
	Temp_F float64 `json:"temp_f"`
}

func GetWeather(city string) (*Temperature, error) {
	// Desabilitar a verificação do certificado SSL
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	config, _ := configs.LoadConfig()

	city = url.QueryEscape(city)
	// log.Println("https://api.weatherapi.com/v1/current.json?key=" + config.WeatherapiKey + "&q=" + city + "&aqi=no")

	resp, error := http.Get("https://api.weatherapi.com/v1/current.json?key=" + config.WeatherapiKey + "&q=" + city + "&aqi=no")
	if error != nil {
		return nil, error
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}

	// fmt.Println("Weather: ", string(body))

	var w Weatherapi
	error = json.Unmarshal(body, &w)
	if error != nil {
		return nil, error
	}

	if w.Current.TempC == 0 {
		return nil, fmt.Errorf("could not retrieve temperature for city: %s", city)
	}

	t := formatTemparature(w.Current.TempC)
	return &t, nil
}

func formatTemparature(celsius float64) Temperature {
	return Temperature{
		Temp_C: celsius,
		Temp_K: celsius + 273,
		Temp_F: celsius*1.8 + 32,
	}
}
