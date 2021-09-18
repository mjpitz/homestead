package geocoding

type Benchmark struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"benchmarkName,omitempty"`
	Description string `json:"benchmarkDescription,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

type Address struct {
	Street string `json:"street,omitempty"`
	City   string `json:"city,omitempty"`
	State  string `json:"state,omitempty"`
	Zip    string `json:"zip,omitempty"`
}

type Coordinates struct {
	X float32 `json:"x,omitempty"`
	Y float32 `json:"y,omitempty"`
}

type TigerLine struct {
	ID   string `json:"tigerLineId,omitempty"`
	Side string `json:"side,omitempty"`
}

type AddressComponents struct {
	FromAddress     string `json:"fromAddress,omitempty"`
	ToAddress       string `json:"toAddress,omitempty"`
	PreQualifier    string `json:"preQualifier,omitempty"`
	PreDirection    string `json:"preDirection,omitempty"`
	PreType         string `json:"preType,omitempty"`
	StreetName      string `json:"streetName,omitempty"`
	SuffixType      string `json:"suffixType,omitempty"`
	SuffixDirection string `json:"suffixDirection,omitempty"`
	SuffixQualifier string `json:"suffixQualifier,omitempty"`
	City            string `json:"city,omitempty"`
	State           string `json:"state,omitempty"`
	Zip             string `json:"zip,omitempty"`
}

type AddressMatch struct {
	MatchedAddress    string             `json:"matchedAddress,omitempty"`
	Coordinates       *Coordinates       `json:"coordinates,omitempty"`
	TigerLine         *TigerLine         `json:"tigerLine,omitempty"`
	AddressComponents *AddressComponents `json:"addressComponents,omitempty"`
}

type SearchInput struct {
	Benchmark *Benchmark `json:"benchmark,omitempty"`
	Address   *Address   `json:"address,omitempty"`
}

type SearchResult struct {
	Input          *SearchInput    `json:"input,omitempty"`
	AddressMatches []*AddressMatch `json:",omitempty"`
}

type SearchByAddressResponse struct {
	Result *SearchResult `json:"result,omitempty"`
}
