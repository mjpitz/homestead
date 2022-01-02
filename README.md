Some tools to help with my homesteading. Since I'm based out of the US, I used existing government APIs to develop this.
I will likely augment some of this with my own measurements, but will initially use existing data.

## Datasets

### weather

The `weather-index-builder` creates/updates a [bleve][] index. Each document is a "reading" from the weather API broken 
down into 15 minute segments. Each reading is marked with an `observed_at` time that allows for multiple readings to
inform a measure for a window. For example, one might use an average, percentile, or combination of both to inform them.

[bleve]: http://blevesearch.com/docs/Home/
