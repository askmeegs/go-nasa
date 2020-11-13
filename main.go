package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
)

func main() {
	fmt.Println("🪐 Starting server...")
	http.HandleFunc("/", Home)
	http.ListenAndServe(":8080", nil)
}

type WeatherOneDay struct {
	Temp           AT                       `json:"AT"`
	WindSpeed      HWS                      `json:"HWS"`
	Pressure       PRE                      `json:"PRE"`
	Season         string                   `json:"Season"`
	WindDirections map[string]WindDirection `json:"WD"`
	EarthDate      string                   `json:"Last_UTC"`
}

// atmospheric temp
type AT struct {
	MinTemp float64 `json:"mn"`
	MaxTemp float64 `json:"mx"`
}

// wind speed
type HWS struct {
	Average float64 `json:"av"`
}

// pressure
type PRE struct {
	Average float64 `json:"av"`
}
type WindDirection struct {
	compassPoint string `json:"compass_point"`
	ct           int    `json:"ct"`
}

type Photo struct {
	Source string `json:"img_src"`
	Date   string `json:"earth_date"`
	Rover  Rover  `json:"rover"`
}

type Rover struct {
	Name string `json:"name"`
}

type PageData struct {
	WeatherConditions WeatherOneDay
	Photos            []Photo
}

// app has one endpoint, "/" which fetches Mars API data and renders it in index.html
func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Attempting to serve homepage.")

	W, err := getWeather()
	if err != nil {
		fmt.Printf("Error getting weather: %v", err)
	}

	// Render template
	Photos, err := getPhotos()
	if err != nil {
		fmt.Printf("Error getting photos: %v", err)
	}

	data := &PageData{W, Photos}

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("rendering HTML template...")
	err = t.Execute(w, data)
	if err != nil { // if there is an error
		fmt.Println(err)
	}
}

// Get current weather conditions on Mars
// https://mars.nasa.gov/insight/weather/
// https://api.nasa.gov/assets/insight/InSight%20Weather%20API%20Documentation.pdf
// https://api.nasa.gov/insight_weather/?api_key=DEMO_KEY&feedtype=json&ver=1.0

func getWeather() (WeatherOneDay, error) {
	apiKey := "HiJkp873RSNocNvf3uUbVSoBavTMqIP7SI7dlW12"
	url := fmt.Sprintf("https://api.nasa.gov/insight_weather/?api_key=%s&feedtype=json&ver=1.0", apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return WeatherOneDay{}, err
	}

	// Read response body into JSON
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return WeatherOneDay{}, err
	}
	// Get raw data
	var rawData map[string]interface{}
	err = json.Unmarshal(body, &rawData)
	if err != nil {
		return WeatherOneDay{}, err
	}

	// keys are a series of sol Strings (eg. 691, 692...) then sol_keys, then validity_checks
	// want to get the first item in the sorted list

	// Get the oldest one-day weather report, most likely to be populated with data, given
	// mars communication lag. ("today" is usually nil)
	keys := make([]string, 0, len(rawData))
	for k := range rawData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println(keys)
	oldestSolIndex := keys[0]
	fmt.Printf("✅ Chose sol: %s\n", oldestSolIndex)

	// parse
	rawWeather := rawData[oldestSolIndex]
	// fmt.Println(rawWeather)
	M, err := json.Marshal(rawWeather)
	if err != nil {
		return WeatherOneDay{}, err
	}
	var output = new(WeatherOneDay)
	err = json.Unmarshal(M, &output)
	if err != nil {
		return WeatherOneDay{}, err
	}

	fmt.Println(output)
	// Done!
	return *output, nil
}

// Gets a random Mars rover photo
// https://api.nasa.gov/mars-photos/api/v1/rovers/curiosity/photos?sol=523&camera=fhaz&api_key=DEMO_KEY
func getPhotos() ([]Photo, error) {
	apiKey := "HiJkp873RSNocNvf3uUbVSoBavTMqIP7SI7dlW12"
	url := fmt.Sprintf("https://api.nasa.gov/mars-photos/api/v1/rovers/curiosity/photos?sol=523&camera=fhaz&api_key=%s", apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return []Photo{}, err
	}

	// Read response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Photo{}, err
	}

	// Unmarshal into map["photos"][photoList]
	var rawData map[string]interface{}
	err = json.Unmarshal(body, &rawData)
	if err != nil {
		return []Photo{}, err
	}

	// Get list of photos out of the value at map["photos"]
	rawPhotos := rawData["photos"]
	M, err := json.Marshal(rawPhotos)
	if err != nil {
		return []Photo{}, err
	}
	var output = []Photo{}
	err = json.Unmarshal(M, &output)
	if err != nil {
		return []Photo{}, err
	}
	// Done!
	return output, nil
}
