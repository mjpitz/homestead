Some tools to help with my homesteading. Since I'm based out of the US, I used existing government APIs to develop this.
I will likely augment some of this with my own measurements, but will initially use existing data.

## Libraries

### `internal/index`

Create, update, and query inverted indexes in [badger][]. While this solution isn't as feature rich as something like
[bleve][], it aims to provide the minimal functionality needed for analyzing data.

## Service

### emquery

`emquery` provides a [Grafana][] [SimpleJSON][] implementation backed by data from `internal/index` datasets.

## Datasets

### weather

The `weather-index-builder` creates/updates an `internal/index`. Each document is a "reading" from the weather API 
broken down into 15 minute segments. Each reading is marked with an `observed_at` time that allows for multiple readings 
to inform a measure for a window. For example, one might use an average, percentile, or combination of both to inform 
them.

[badger]: https://dgraph.io/docs/badger/
[Grafana]: https://grafana.com/oss/grafana/
[SimpleJSON]: https://grafana.com/grafana/plugins/simpod-json-datasource/
