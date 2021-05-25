package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"strconv"
	"net/http"
	"time"
)

type Sensebox struct {
	ID                string          `json:"_id"`
	Createdat         time.Time       `json:"createdAt"`
	Updatedat         time.Time       `json:"updatedAt"`
	Name              string          `json:"name"`
	Currentlocation   Currentlocation `json:"currentLocation"`
	Grouptag          string          `json:"grouptag"`
	Exposure          string          `json:"exposure"`
	Sensors           []Sensors       `json:"sensors"`
	Model             string          `json:"model"`
	Image             string          `json:"image"`
	Lastmeasurementat time.Time       `json:"lastMeasurementAt"`
	Loc               []Loc           `json:"loc"`
}
type Currentlocation struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
	Timestamp   time.Time `json:"timestamp"`
}
type Lastmeasurement struct {
	Value     string    `json:"value"`
	Createdat time.Time `json:"createdAt"`
}
type Sensors struct {
	Title           string          `json:"title"`
	Unit            string          `json:"unit"`
	Sensortype      string          `json:"sensorType"`
	Icon            string          `json:"icon"`
	ID              string          `json:"_id"`
	Lastmeasurement Lastmeasurement `json:"lastMeasurement"`
}
type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
	Timestamp   time.Time `json:"timestamp"`
}
type Loc struct {
	Geometry Geometry `json:"geometry"`
	Type     string   `json:"type"`
}

var (
	senseboxUrl = "https://api.opensensemap.org/boxes/5a0c30489fd3c20011115fb7"
)

func GetSensebox(senseboxUrl string) Sensebox {
	resp, err := http.Get(senseboxUrl)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	var s Sensebox

	json.NewDecoder(resp.Body).Decode(&s)
	return s
}

func recordMetrics() {
	go func() {
		for {
			sensebox := GetSensebox(senseboxUrl)
			fmt.Println(sensebox.Model)
			for _, sensor := range sensebox.Sensors {
				fmt.Printf("Sensor %s returned %s\n", sensor.Title, sensor.Lastmeasurement.Value)

				phenomenon, err := strconv.ParseFloat(sensor.Lastmeasurement.Value, 32)
				if err != nil {
					log.Printf("Invalid %s value: %s\n", sensor.Title, sensor.Lastmeasurement.Value)
				}
				gauge := gaugeMap[sensor.Title]
				if gauge != nil {
					gauge.Set(phenomenon)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	tempGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sensebox_temperature",
		Help: "The temperature of the sensebox",
	})
)

func makeGauge(gaugeName string) prometheus.Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sensebox_" + gaugeName,
		Help: "The " + gaugeName,
	})
}

var (
	gaugeMap = map[string]prometheus.Gauge{
		"Temperatur":         tempGauge,
		"Luftdruck":          makeGauge("luftdruck"),
		"Beleuchtungsstärke": makeGauge("beleuchtungsstaerke"),
		"rel. Luftfeuchte":   makeGauge("relluftfeuchte"),
		"UV-Intensität":      makeGauge("uvintensity"),
		"PM10":               makeGauge("pm10"),
		"PM2.5":              makeGauge("pm2dot5"),
	}
)

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9500", nil))
}
