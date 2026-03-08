package weather

import "time"

// Data represents a weather snapshot.
type Data struct {
	TempCelsius float64   `json:"temp_celsius"`
	Condition   string    `json:"condition"`
	Description string    `json:"description"`
	Humidity    int       `json:"humidity"`
	IconCode    string    `json:"icon_code"`
	FetchedAt   time.Time `json:"fetched_at"`
}
