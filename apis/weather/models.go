package weather

type Measurement struct {
	ValidTime string  `json:"validTime,omitempty"`
	Value     float32 `json:"value,omitempty"`
}

type Elevation struct {
	UnitCode string  `json:"unitCode,omitempty"`
	Value    float32 `json:"value,omitempty"`
}

type DataPoints struct {
	UnitOfMeasure string         `json:"uom,omitempty"`
	Values        []*Measurement `json:"values,omitempty"`
}

type Forecast struct {
	StartTime        string `json:"startTime,omitempty"`
	EndTime          string `json:"endTime,omitempty"`
	Temperature      int    `json:"temperature,omitempty"`
	TemperatureUnit  string `json:"temperatureUnit,omitempty"`
	WindSpeed        string `json:"windSpeed,omitempty"`
	WindDirection    string `json:"windDirection,omitempty"`
	ShortForecast    string `json:"shortForecast,omitempty"`
	DetailedForecast string `json:"detailedForecast,omitempty"`
}

type Coordinates struct {
	Latitude  float32 `json:"latitude,omitempty"`
	Longitude float32 `json:"longitude,omitempty"`
}

type PointProperties struct {
	GridID       string `json:"gridId,omitempty"`
	GridX        int    `json:"gridX,omitempty"`
	GridY        int    `json:"gridY,omitempty"`
	TimeZone     string `json:"timeZone,omitempty"`
	RadarStation string `json:"radarStation,omitempty"`
}

type GridpointProperties struct {
	UpdateTime                       string      `json:"updateTime,omitempty"`
	ValidTimes                       string      `json:"validTimes,omitempty"`
	Elevation                        *Elevation  `json:"elevation,omitempty"`
	GridID                           string      `json:"gridId,omitempty"`
	GridX                            string      `json:"gridX,omitempty"`
	GridY                            string      `json:"gridY,omitempty"`
	Temperature                      *DataPoints `json:"temperature,omitempty"`
	Dewpoint                         *DataPoints `json:"dewpoint,omitempty"`
	MaxTemperature                   *DataPoints `json:"maxTemperature,omitempty"`
	MinTemperature                   *DataPoints `json:"minTemperature,omitempty"`
	RelativeHumidity                 *DataPoints `json:"relativeHumidity,omitempty"`
	ApparentTemperature              *DataPoints `json:"apparentTemperature,omitempty"`
	HeatIndex                        *DataPoints `json:"heatIndex,omitempty"`
	WindChill                        *DataPoints `json:"windChill,omitempty"`
	SkyCover                         *DataPoints `json:"skyCover,omitempty"`
	WindDirection                    *DataPoints `json:"windDirection,omitempty"`
	WindSpeed                        *DataPoints `json:"windSpeed,omitempty"`
	WindGust                         *DataPoints `json:"windGust,omitempty"`
	Hazards                          *DataPoints `json:"hazards,omitempty"`
	ProbabilityOfPrecipitation       *DataPoints `json:"probabilityOfPrecipitation,omitempty"`
	QuantitativePrecipitation        *DataPoints `json:"quantitativePrecipitation,omitempty"`
	IceAccumulation                  *DataPoints `json:"iceAccumulation,omitempty"`
	SnowfallAmount                   *DataPoints `json:"snowfallAmount,omitempty"`
	SnowLevel                        *DataPoints `json:"snowLevel,omitempty"`
	CeilingHeight                    *DataPoints `json:"ceilingHeight,omitempty"`
	Visibility                       *DataPoints `json:"visibility,omitempty"`
	TransportWindSpeed               *DataPoints `json:"transportWindSpeed,omitempty"`
	TransportWindDirection           *DataPoints `json:"transportWindDirection,omitempty"`
	MixingHeight                     *DataPoints `json:"mixingHeight,omitempty"`
	HainesIndex                      *DataPoints `json:"hainesIndex,omitempty"`
	LightningActivityLevel           *DataPoints `json:"lightningActivityLevel,omitempty"`
	TwentyFootWindSpeed              *DataPoints `json:"twentyFootWindSpeed,omitempty"`
	TwentyFootWindDirection          *DataPoints `json:"twentyFootWindDirection,omitempty"`
	WaveHeight                       *DataPoints `json:"waveHeight,omitempty"`
	WavePeriod                       *DataPoints `json:"wavePeriod,omitempty"`
	PrimarySwellHeight               *DataPoints `json:"primarySwellHeight,omitempty"`
	PrimarySwellDirection            *DataPoints `json:"primarySwellDirection,omitempty"`
	SecondarySwellHeight             *DataPoints `json:"secondarySwellHeight,omitempty"`
	SecondarySwellDirection          *DataPoints `json:"secondarySwellDirection,omitempty"`
	WavePeriod2                      *DataPoints `json:"wavePeriod2,omitempty"`
	WindWaveHeight                   *DataPoints `json:"windWaveHeight,omitempty"`
	DispersionIndex                  *DataPoints `json:"dispersionIndex,omitempty"`
	Pressure                         *DataPoints `json:"pressure,omitempty"`
	ProbabilityOfTropicalStormWinds  *DataPoints `json:"probabilityOfTropicalStormWinds,omitempty"`
	ProbabilityOfHurricaneWinds      *DataPoints `json:"probabilityOfHurricaneWinds,omitempty"`
	PotentialOf15mphWinds            *DataPoints `json:"potentialOf15mphWinds,omitempty"`
	PotentialOf25mphWinds            *DataPoints `json:"potentialOf25mphWinds,omitempty"`
	PotentialOf35mphWinds            *DataPoints `json:"potentialOf35mphWinds,omitempty"`
	PotentialOf45mphWinds            *DataPoints `json:"potentialOf45mphWinds,omitempty"`
	PotentialOf20mphWindGusts        *DataPoints `json:"potentialOf20mphWindGusts,omitempty"`
	PotentialOf30mphWindGusts        *DataPoints `json:"potentialOf30mphWindGusts,omitempty"`
	PotentialOf40mphWindGusts        *DataPoints `json:"potentialOf40mphWindGusts,omitempty"`
	PotentialOf50mphWindGusts        *DataPoints `json:"potentialOf50mphWindGusts,omitempty"`
	PotentialOf60mphWindGusts        *DataPoints `json:"potentialOf60mphWindGusts,omitempty"`
	GrasslandFireDangerIndex         *DataPoints `json:"grasslandFireDangerIndex,omitempty"`
	ProbabilityOfThunder             *DataPoints `json:"probabilityOfThunder,omitempty"`
	DavisStabilityIndex              *DataPoints `json:"davisStabilityIndex,omitempty"`
	AtmosphericDispersionIndex       *DataPoints `json:"atmosphericDispersionIndex,omitempty"`
	LowVisibilityOccurrenceRiskIndex *DataPoints `json:"lowVisibilityOccurrenceRiskIndex,omitempty"`
	Stability                        *DataPoints `json:"stability,omitempty"`
	RedFlagThreatIndex               *DataPoints `json:"redFlagThreatIndex,omitempty"`
}

type ForecastProperties struct {
	UpdateTime string      `json:"updateTime,omitempty"`
	ValidTimes string      `json:"validTimes,omitempty"`
	Elevation  *Elevation  `json:"elevation,omitempty"`
	Periods    []*Forecast `json:"periods,omitempty"`
}

type Response struct {
	Properties interface{} `json:"properties,omitempty"`
}
