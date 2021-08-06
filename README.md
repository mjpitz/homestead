Some tools to help with my homesteading. Since I'm based out of the US, I used open data APIs to develop this. 

## Geocoding

The US Census provides a Geocoding API that takes a street address, a benchmark, and obtain geo-coordinates. We use the
coordinates for various other API calls such as fetching the weather forecast and history data. 

- https://geocoding.geo.census.gov/geocoder/benchmarks
  - ?format=json
  - 4 = current range
  - 2020 = census range
  
- https://geocoding.geo.census.gov/geocoder/locations/address
  - ?format=json
  - &street=
  - &city=
  - &state=
  - &zip=
  - &benchmark=2020

## Weather

weather.gov provides an API, however you mush work with geo-coordinates. First we will need to obtain the `gridID`,
`gridX`, and `gridY`. Then, once you get these pieces you can get forecasts for the area. You can also get station
information too. Not sure if you can get raw data streams from stations directly.

- https://api.weather.gov/points/{lat},{long}
- https://api.weather.gov/gridpoints/{gridID}/{gridX},{grixY}/forecast/hourly
- https://api.weather.gov/gridpoints/{gridID}/{gridX},{girdY}/forecast
