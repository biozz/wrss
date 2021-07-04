package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/gorilla/feeds"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	env       = flag.Bool("env", false, "display env variables")
	feedsPath = flag.String("feeds", "feeds.yml", "path to feeds.yml config")
	yc        *YandexWeatherClient
	c         Config
	fc        FeedsConfig
	cache     *bigcache.BigCache
	log       = logrus.New()
)

type Config struct {
	ServerUrl     string         `default:"http://localhost:8090"`
	Addr          string         `default:":8090" split_words:"true"`
	YandexWeather YanddexWeather `split_words:"true"`
}

type YanddexWeather struct {
	APIKey  string `default:"change_me" split_words:"true"`
	BaseURL string `default:"https://api.weather.yandex.ru" split_words:"true"`
}

type Feed struct {
	Name string `yaml:"name"`
	Slug string `yaml:"slug"`
	Lat  string `yaml:"lat"`
	Lon  string `yaml:"lon"`
	Lang string `yaml:"lang"`
}

type FeedsConfig struct {
	Feeds []Feed `yaml:",flow"`
}

func main() {
	flag.Parse()

	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	if *env {
		err = envconfig.Usage("", &c)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Out = os.Stdout

	cache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(24 * 60 * time.Minute))

	yamlFile, err := ioutil.ReadFile(*feedsPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &fc)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	yc = NewYandexWeatherClient(c.YandexWeather.BaseURL, c.YandexWeather.APIKey)

	r := mux.NewRouter()
	r.HandleFunc("/feed", feed)

	loggedRouter := handlers.LoggingHandler(log.Writer(), r)
	http.Handle("/", loggedRouter)

	log.Printf("Starting server at %s\n", c.Addr)
	log.Fatal(http.ListenAndServe(c.Addr, nil))
}

func feed(rw http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	feedTextCached, _ := cache.Get(slug)
	if feedText := string(feedTextCached); feedText != "" {
		log.Printf("Cache hit for %s", slug)
		fmt.Fprint(rw, feedText)
		return
	}
	feedConfig := getFeedConfigBySlug(slug)
	if feedConfig == nil {
		msg := fmt.Sprintf("No config with slug %s", slug)
		log.Println(msg)
		fmt.Fprint(rw, msg)
		return
	}
	res, err := yc.GetForecast(feedConfig.Lat, feedConfig.Lon, feedConfig.Lang, true)
	if err != nil {
		log.Fatal(err.Error())
	}
	feedText, err := generateFeedFromYandexWeatherResponse(slug, res)
	if err != nil {
		msg := fmt.Sprintf("Unable to generate feed %v", err)
		log.Println(err)
		fmt.Fprint(rw, msg)
		return
	}
	cache.Set(slug, []byte(feedText))
	fmt.Fprint(rw, feedText)
}

func getFeedConfigBySlug(slug string) *Feed {
	for _, feed := range fc.Feeds {
		if feed.Slug != slug {
			continue
		}
		return &feed
	}
	return nil
}

func generateFeedFromYandexWeatherResponse(slug string, ywr *YandexWeatherResponse) (string, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Weather RSS",
		Link:        &feeds.Link{Href: fmt.Sprintf("%s/feeds?slug=%s", c.ServerUrl, slug)},
		Description: "Weather forecast delivered as an RSS feed",
		Created:     now,
	}
	const tpl = `<p>{{.Fact.Temp}}ºC</p>`
	t, err := template.New("content").Parse(tpl)
	if err != nil {
		return "", err
	}
	var tplBuffer bytes.Buffer
	err = t.Execute(&tplBuffer, ywr)
	if err != nil {
		return "", err
	}
	feed.Items = []*feeds.Item{
		{
			Title:       fmt.Sprintf("%s - погода на %s", ywr.GeoObject.Locality.Name, ywr.NowDT),
			Link:        &feeds.Link{Href: ywr.Info.URL},
			Description: "Прогоноз погоды",
			Author:      &feeds.Author{Name: "Yandex"},
			Created:     now,
			Content:     tplBuffer.String(),
		},
	}
	atom, err := feed.ToAtom()
	if err != nil {
		return "", err
	}
	return atom, nil
}
