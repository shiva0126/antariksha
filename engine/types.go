package engine

import "time"

const SchemaVersion = 3

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
	TZ  string  `json:"tz"`
}

type Limb struct {
	Name   string `json:"name"`
	Number int    `json:"number,omitempty"`
	EndsAt string `json:"ends_at"`
}

type Window struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
type Choghadiya struct {
	Name  string `json:"name"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type Day struct {
	SchemaVersion int          `json:"schema_version"`
	Date          string       `json:"date"`
	Location      Location     `json:"location"`
	Sunrise       string       `json:"sunrise"`
	Sunset        string       `json:"sunset"`
	Moonrise      string       `json:"moonrise,omitempty"`
	Moonset       string       `json:"moonset,omitempty"`
	Vaara         string       `json:"vaara"`
	Paksha        string       `json:"paksha"`
	LunarMonth    string       `json:"lunar_month"`
	Tithi         Limb         `json:"tithi"`
	Nakshatra     Limb         `json:"nakshatra"`
	Yoga          Limb         `json:"yoga"`
	Karana        Limb         `json:"karana"`
	RahuKaal      Window       `json:"rahu_kaal"`
	Yamaganda     Window       `json:"yamaganda"`
	Gulika        Window       `json:"gulika"`
	Abhijit       Window       `json:"abhijit"`
	Choghadiya    []Choghadiya `json:"choghadiya"`
	Festivals     []string     `json:"festivals"`
}

type Computed struct {
	Day             Day
	Sunrise, Sunset time.Time
}
