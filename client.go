package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type YandexWeatherClient struct {
	URL    string
	APIKey string
}

func NewYandexWeatherClient(baseURL, yandexWeatherAPIKey string) *YandexWeatherClient {
	return &YandexWeatherClient{
		URL:    baseURL,
		APIKey: yandexWeatherAPIKey,
	}
}

type YandexWeatherResponse struct {
	Now       int32     `json:"now"`
	NowDT     string    `json:"now_dt"`
	Info      Info      `json:"info"`
	Fact      Fact      `json:"fact"`
	GeoObject GeoObject `json:"geo_object"`
}

type Info struct {
	URL string `json:"url"`
}

type Fact struct {
	Temp int `json:"temp"`
}

type GeoObject struct {
	Locality GeoItem `json:"locality"`
}

type GeoItem struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

func (y *YandexWeatherClient) GetForecast(lat, lon, lang string, extra bool) (*YandexWeatherResponse, error) {
	url := fmt.Sprintf("%s/v2/forecast/", y.URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Yandex-API-Key", y.APIKey)
	q := req.URL.Query()
	q.Add("lat", lat)
	q.Add("lon", lon)
	q.Add("lang", lang)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var weatherResponse YandexWeatherResponse
	err = decoder.Decode(&weatherResponse)
	if err != nil {
		return nil, err
	}
	return &weatherResponse, nil
}
