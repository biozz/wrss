# weather-rss

RSS is great, why not use it for weather forecasts? This project is just about that.

It is a web server written in Go. You define places in `feeds.yml` and then they are accessible through `/feed?slug=rybinsk` endpoint which generates Atom feed with the data from [Yandex.Weather](https://yandex.com/weather). Each forecast is cached for 24 hours, because of the strict limits of the free tier Yandex.Weather API.

TODO:

- add more notes about usage
