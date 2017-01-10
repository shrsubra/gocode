package main
import (
    "net/http"
    "encoding/json"
    "strings"
    "log"
)

type weatherProvider interface {
    temperature(city string) (float64, error)
}

type multiWeatherProvider []weatherProvider

type openWeatherMap struct {
    ApiKey string
}
type weatherUnderground struct {
    ApiKey string
}

func (provider openWeatherMap)temperature(city string) (float64, error){
    resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + provider.ApiKey + "&q=" + city)
    if err != nil{
        return 0, err
    }
    defer resp.Body.Close()
    var d struct{
        Main struct {
            Kelvin float64 `json:"temp"`
        } `json:"main"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil{
        return 0, err
    }
    log.Printf("Openweathermap temperature for %s in kelvin is: %.2f", city, d.Main.Kelvin)
    return d.Main.Kelvin, nil
}

func (provider weatherUnderground) temperature(city string) (float64, error) {
    resp,err := http.Get("http://api.wunderground.com/api/" + provider.ApiKey + "/conditions/q/CA/" + city + ".json")
    if err != nil{
        return 0, err
    }
    defer resp.Body.Close()
    var d struct {
        Observation struct{
            Celsius float64 `json:"temp_c"`
        } `json:"current_observation"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil{
        return 0, err
    }
    log.Printf("WeatherUnderground temperature for %s in celsius is: %.2f", city, d.Observation.Celsius)
    return d.Observation.Celsius + 273.15, nil
}

func (providers multiWeatherProvider) temperature(city string) (float64, error) {
    sum := 0.0
    for _, provider := range providers{
        k, err := provider.temperature(city)
        if err != nil {
            return 0, err
        }
        sum += k
    }
    return sum/float64(len(providers)), nil
}

func main() {
    mw := multiWeatherProvider{
        openWeatherMap{ApiKey: ""},
        weatherUnderground{ApiKey:""},
    }
    http.HandleFunc("/hello", hello)
    http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request){
        city := strings.SplitN(r.URL.Path, "/", 3)[2]
        d,err := mw.temperature(city)
        if err != nil{
            http.Error(w, err.Error() + "aa", http.StatusInternalServerError)
            return
        } 
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        json.NewEncoder(w).Encode(d)
               
    })
    http.ListenAndServe(":8080", nil)
}

func hello(w http.ResponseWriter, r *http.Request){
    w.Write([]byte("hello, the weather for your city is\n"))
}

func query(city string) (weatherData, error){
    resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=242c715343d9d960a9d528222480bfbc&q=" + city)
    if err != nil{
        return weatherData{}, err
    }
    defer resp.Body.Close()
    var d weatherData
    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
        return weatherData{}, err
    }
    return d, nil
}
type weatherData struct {
    Name string `json:"name"`
    Main struct {
        Kelvin float64 `json:"temp"`
    } `json:"main"`
}
