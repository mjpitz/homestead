### Weather

weather.gov provides an API, however you mush work with geo-coordinates. First we will need to obtain the `gridID`,
`gridX`, and `gridY`. Then, once you get these pieces you can get forecasts for the area. You can also get station
information too. Not sure if you can get raw data streams from stations directly.

- https://www.weather.gov/documentation/services-web-api
- https://api.weather.gov/points/{lat},{long}
- https://api.weather.gov/gridpoints/{gridID}/{gridX},{gridY}
- https://api.weather.gov/gridpoints/{gridID}/{gridX},{gridY}/forecast/hourly
- https://api.weather.gov/gridpoints/{gridID}/{gridX},{girdY}/forecast
- https://api.weather.gov/icons

The gridpoints endpoint provides a significant amount of data. Forecasts seem to be built off their own models.
