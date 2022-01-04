package main

import (
	"context"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/geo"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/homestead/internal/apis/geocoding"
	"github.com/mjpitz/homestead/internal/apis/weather"
	"github.com/mjpitz/homestead/internal/index"
	"github.com/mjpitz/myago/clocks"
	"github.com/mjpitz/myago/config"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
)

type Weather struct {
	Timestamp  time.Time `json:"timestamp"`
	ObservedAt time.Time `json:"observed_at"`

	Coordinates                      geo.Point `json:"coordinates"`
	Elevation                        float64   `json:"elevation_m"`
	Temperature                      float64   `json:"temperature_degc"`
	Dewpoint                         float64   `json:"dewpoint_degc"`
	MaxTemperature                   float64   `json:"max_temperature_degc"`
	MinTemperature                   float64   `json:"min_temperature_degc"`
	RelativeHumidity                 float64   `json:"relative_humidity_pct"`
	ApparentTemperature              float64   `json:"apparent_temperature_degc"`
	HeatIndex                        float64   `json:"heat_index_degc"`
	WindChill                        float64   `json:"wind_chill_degc"`
	SkyCover                         float64   `json:"sky_cover_pct"`
	WindDirection                    float64   `json:"wind_direction"`
	WindSpeed                        float64   `json:"wind_speed_kph"` // kilometers per hour
	WindGust                         float64   `json:"wind_gust_kph"`  // kilometers per hour
	ProbabilityOfPrecipitation       float64   `json:"precipitation_probability_pct"`
	QuantitativePrecipitation        float64   `json:"precipitation_quantity_mm"`
	IceAccumulation                  float64   `json:"ice_accumulation_mm"`
	SnowfallAmount                   float64   `json:"snowfall_amount_mm"`
	SnowLevel                        float64   `json:"snow_level"`
	CeilingHeight                    float64   `json:"ceiling_height"`
	Visibility                       float64   `json:"visibility"`
	TransportWindSpeed               float64   `json:"transport_wind_speed_kph"`
	TransportWindDirection           float64   `json:"transport_wind_direction"`
	MixingHeight                     float64   `json:"mixing_height_m"`
	HainesIndex                      float64   `json:"haines_index"`
	LightningActivityLevel           float64   `json:"lightning_activity_level"`
	TwentyFootWindSpeed              float64   `json:"twenty_foot_wind_speed_kph"`
	TwentyFootWindDirection          float64   `json:"twenty_foot_wind_direction"`
	WaveHeight                       float64   `json:"wave_height"`
	WavePeriod                       float64   `json:"wave_period"`
	PrimarySwellHeight               float64   `json:"primary_swell_height"`
	PrimarySwellDirection            float64   `json:"primary_swell_direction"`
	SecondarySwellHeight             float64   `json:"secondary_swell_height"`
	SecondarySwellDirection          float64   `json:"secondary_swell_direction"`
	WavePeriod2                      float64   `json:"wave_period_2"`
	WindWaveHeight                   float64   `json:"wind_wave_height"`
	DispersionIndex                  float64   `json:"dispersion_index"`
	Pressure                         float64   `json:"pressure"`
	ProbabilityOfTropicalStormWinds  float64   `json:"probability_of_tropical_storm_winds"`
	ProbabilityOfHurricaneWinds      float64   `json:"probability_of_hurricane_winds"`
	PotentialOf15mphWinds            float64   `json:"potential_of_15_mph_winds"`
	PotentialOf25mphWinds            float64   `json:"potential_of_25_mph_winds"`
	PotentialOf35mphWinds            float64   `json:"potential_of_35_mph_winds"`
	PotentialOf45mphWinds            float64   `json:"potential_of_45_mph_winds"`
	PotentialOf20mphWindGusts        float64   `json:"potential_of_20_mph_wind_gusts"`
	PotentialOf30mphWindGusts        float64   `json:"potential_of_30_mph_wind_gusts"`
	PotentialOf40mphWindGusts        float64   `json:"potential_of_40_mph_wind_gusts"`
	PotentialOf50mphWindGusts        float64   `json:"potential_of_50_mph_wind_gusts"`
	PotentialOf60mphWindGusts        float64   `json:"potential_of_60_mph_wind_gusts"`
	GrasslandFireDangerIndex         float64   `json:"grassland_fire_danger_index"`
	ProbabilityOfThunder             float64   `json:"probability_of_thunder"`
	DavisStabilityIndex              float64   `json:"davis_stability_index"`
	AtmosphericDispersionIndex       float64   `json:"atmospheric_dispersion_index"`
	LowVisibilityOccurrenceRiskIndex float64   `json:"low_visibility_occurrence_risk_index"`
	Stability                        float64   `json:"stability"`
	RedFlagThreatIndex               float64   `json:"red_flag_threat_index"`
}

func (w Weather) Type() string {
	return "weather"
}

var _ mapping.Classifier = &Weather{}

type Config struct {
	ConfigFile string            `json:"config_file" usage:"specify the location of a file containing the configuration"`
	Index      index.Config      `json:"index"`
	Address    geocoding.Address `json:"address"`
	Log        zaputil.Config    `json:"log"`
}

var docFrequency = 15 * time.Minute

func update(idx map[time.Time]*Weather, points *weather.DataPoints, set func(w *Weather, value float64)) {
	for _, measure := range points.Values {
		t := measure.ValidTime.Time
		d := measure.ValidTime.Duration

		if _, ok := idx[t]; !ok {
			idx[t] = &Weather{}
		}

		if d == 0 {
			// no duration, single document
			set(idx[t], float64(measure.Value))
			return
		}

		// expand window

		for d > 0 {
			if _, ok := idx[t]; !ok {
				idx[t] = &Weather{}
			}

			set(idx[t], float64(measure.Value))

			t = t.Add(docFrequency)
			d = d - docFrequency
		}
	}
}

func main() {
	cfg := &Config{
		Index: index.Config{
			Tags: cli.NewStringSlice(),
		},
	}

	app := &cli.App{
		Name:      "weather-index-builder",
		Usage:     "Construct an index with weather related data.",
		UsageText: "weather-index-builder [options]",
		Flags:     flagset.Extract(cfg),
		Before: func(ctx *cli.Context) (err error) {
			ctx.Context = zaputil.Setup(ctx.Context, cfg.Log)

			if cfg.ConfigFile != "" {
				err := config.Load(ctx.Context, cfg, cfg.ConfigFile)
				if err != nil {
					return err
				}
			}

			return err
		},
		Action: func(ctx *cli.Context) error {
			builder := index.Builder{
				Action: func(ctx context.Context, index bleve.Index) error {
					geocodingAPI := geocoding.NewClient()
					weatherAPI := weather.NewClient()

					geocodeResp, err := geocodingAPI.SearchByAddress(ctx, &cfg.Address)
					if err != nil {
						return err
					}

					coordinates := geocodeResp.Result.AddressMatches[0].Coordinates

					point, err := weatherAPI.GetPoint(ctx, coordinates.Y, coordinates.X)
					if err != nil {
						return err
					}

					gridpoints, err := weatherAPI.GetGridpoint(ctx, point.GridID, point.GridX, point.GridY)
					if err != nil {
						return err
					}

					idx := make(map[time.Time]*Weather)

					// the following block was code-generated from the following command
					// cat ./internal/apis/weather/models.go | grep '*DataPoints' | awk '{print $1}' | xargs -I^ echo 'update(idx, gridpoints.^, func(w *Weather, v float64) { w.^ = v })' | pbcopy
					update(idx, gridpoints.Temperature, func(w *Weather, v float64) { w.Temperature = v })
					update(idx, gridpoints.Dewpoint, func(w *Weather, v float64) { w.Dewpoint = v })
					update(idx, gridpoints.MaxTemperature, func(w *Weather, v float64) { w.MaxTemperature = v })
					update(idx, gridpoints.MinTemperature, func(w *Weather, v float64) { w.MinTemperature = v })
					update(idx, gridpoints.RelativeHumidity, func(w *Weather, v float64) { w.RelativeHumidity = v })
					update(idx, gridpoints.ApparentTemperature, func(w *Weather, v float64) { w.ApparentTemperature = v })
					update(idx, gridpoints.HeatIndex, func(w *Weather, v float64) { w.HeatIndex = v })
					update(idx, gridpoints.WindChill, func(w *Weather, v float64) { w.WindChill = v })
					update(idx, gridpoints.SkyCover, func(w *Weather, v float64) { w.SkyCover = v })
					update(idx, gridpoints.WindDirection, func(w *Weather, v float64) { w.WindDirection = v })
					update(idx, gridpoints.WindSpeed, func(w *Weather, v float64) { w.WindSpeed = v })
					update(idx, gridpoints.WindGust, func(w *Weather, v float64) { w.WindGust = v })
					update(idx, gridpoints.ProbabilityOfPrecipitation, func(w *Weather, v float64) { w.ProbabilityOfPrecipitation = v })
					update(idx, gridpoints.QuantitativePrecipitation, func(w *Weather, v float64) { w.QuantitativePrecipitation = v })
					update(idx, gridpoints.IceAccumulation, func(w *Weather, v float64) { w.IceAccumulation = v })
					update(idx, gridpoints.SnowfallAmount, func(w *Weather, v float64) { w.SnowfallAmount = v })
					update(idx, gridpoints.SnowLevel, func(w *Weather, v float64) { w.SnowLevel = v })
					update(idx, gridpoints.CeilingHeight, func(w *Weather, v float64) { w.CeilingHeight = v })
					update(idx, gridpoints.Visibility, func(w *Weather, v float64) { w.Visibility = v })
					update(idx, gridpoints.TransportWindSpeed, func(w *Weather, v float64) { w.TransportWindSpeed = v })
					update(idx, gridpoints.TransportWindDirection, func(w *Weather, v float64) { w.TransportWindDirection = v })
					update(idx, gridpoints.MixingHeight, func(w *Weather, v float64) { w.MixingHeight = v })
					update(idx, gridpoints.HainesIndex, func(w *Weather, v float64) { w.HainesIndex = v })
					update(idx, gridpoints.LightningActivityLevel, func(w *Weather, v float64) { w.LightningActivityLevel = v })
					update(idx, gridpoints.TwentyFootWindSpeed, func(w *Weather, v float64) { w.TwentyFootWindSpeed = v })
					update(idx, gridpoints.TwentyFootWindDirection, func(w *Weather, v float64) { w.TwentyFootWindDirection = v })
					update(idx, gridpoints.WaveHeight, func(w *Weather, v float64) { w.WaveHeight = v })
					update(idx, gridpoints.WavePeriod, func(w *Weather, v float64) { w.WavePeriod = v })
					update(idx, gridpoints.PrimarySwellHeight, func(w *Weather, v float64) { w.PrimarySwellHeight = v })
					update(idx, gridpoints.PrimarySwellDirection, func(w *Weather, v float64) { w.PrimarySwellDirection = v })
					update(idx, gridpoints.SecondarySwellHeight, func(w *Weather, v float64) { w.SecondarySwellHeight = v })
					update(idx, gridpoints.SecondarySwellDirection, func(w *Weather, v float64) { w.SecondarySwellDirection = v })
					update(idx, gridpoints.WavePeriod2, func(w *Weather, v float64) { w.WavePeriod2 = v })
					update(idx, gridpoints.WindWaveHeight, func(w *Weather, v float64) { w.WindWaveHeight = v })
					update(idx, gridpoints.DispersionIndex, func(w *Weather, v float64) { w.DispersionIndex = v })
					update(idx, gridpoints.Pressure, func(w *Weather, v float64) { w.Pressure = v })
					update(idx, gridpoints.ProbabilityOfTropicalStormWinds, func(w *Weather, v float64) { w.ProbabilityOfTropicalStormWinds = v })
					update(idx, gridpoints.ProbabilityOfHurricaneWinds, func(w *Weather, v float64) { w.ProbabilityOfHurricaneWinds = v })
					update(idx, gridpoints.PotentialOf15mphWinds, func(w *Weather, v float64) { w.PotentialOf15mphWinds = v })
					update(idx, gridpoints.PotentialOf25mphWinds, func(w *Weather, v float64) { w.PotentialOf25mphWinds = v })
					update(idx, gridpoints.PotentialOf35mphWinds, func(w *Weather, v float64) { w.PotentialOf35mphWinds = v })
					update(idx, gridpoints.PotentialOf45mphWinds, func(w *Weather, v float64) { w.PotentialOf45mphWinds = v })
					update(idx, gridpoints.PotentialOf20mphWindGusts, func(w *Weather, v float64) { w.PotentialOf20mphWindGusts = v })
					update(idx, gridpoints.PotentialOf30mphWindGusts, func(w *Weather, v float64) { w.PotentialOf30mphWindGusts = v })
					update(idx, gridpoints.PotentialOf40mphWindGusts, func(w *Weather, v float64) { w.PotentialOf40mphWindGusts = v })
					update(idx, gridpoints.PotentialOf50mphWindGusts, func(w *Weather, v float64) { w.PotentialOf50mphWindGusts = v })
					update(idx, gridpoints.PotentialOf60mphWindGusts, func(w *Weather, v float64) { w.PotentialOf60mphWindGusts = v })
					update(idx, gridpoints.GrasslandFireDangerIndex, func(w *Weather, v float64) { w.GrasslandFireDangerIndex = v })
					update(idx, gridpoints.ProbabilityOfThunder, func(w *Weather, v float64) { w.ProbabilityOfThunder = v })
					update(idx, gridpoints.DavisStabilityIndex, func(w *Weather, v float64) { w.DavisStabilityIndex = v })
					update(idx, gridpoints.AtmosphericDispersionIndex, func(w *Weather, v float64) { w.AtmosphericDispersionIndex = v })
					update(idx, gridpoints.LowVisibilityOccurrenceRiskIndex, func(w *Weather, v float64) { w.LowVisibilityOccurrenceRiskIndex = v })
					update(idx, gridpoints.Stability, func(w *Weather, v float64) { w.Stability = v })
					update(idx, gridpoints.RedFlagThreatIndex, func(w *Weather, v float64) { w.RedFlagThreatIndex = v })

					observedAt := clocks.Extract(ctx).Now()
					for timestamp, doc := range idx {
						doc.Timestamp = timestamp
						doc.ObservedAt = observedAt
						doc.Elevation = float64(gridpoints.Elevation.Value)
						doc.Coordinates.Lat = float64(coordinates.Y)
						doc.Coordinates.Lon = float64(coordinates.X)

						id := timestamp.Format(time.RFC3339) + "/" + observedAt.Format(time.RFC3339)

						err = index.Index(id, *doc)
						if err != nil {
							return err
						}
					}

					return nil
				},
			}

			return builder.Run(ctx.Context, cfg.Index)
		},
		HideVersion:          true,
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Metadata: map[string]interface{}{
			"arch":       runtime.GOARCH,
			"go_version": strings.TrimPrefix(runtime.Version(), "go"),
			"os":         runtime.GOOS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
