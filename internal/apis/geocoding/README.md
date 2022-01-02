### Geocoding API

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
