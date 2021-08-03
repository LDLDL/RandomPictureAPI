package main

// ArtistsConfItem artists.json bind item
type ArtistsConfItem struct {
	Dir  string	`json:"dir"`
	Name string `json:"name"`
	PID  string `json:"pid"`
}

// ArtistsConf artists.json bind
type ArtistsConf []*ArtistsConfItem

// ArtistInfo artist information
type ArtistInfo struct {
	Name  string `json:"name"`
	PID   string `json:"pid"`
	Count int    `json:"count"`
}

type ArtistsInfo []*ArtistInfo

// Artist with pics
type Artist struct {
	ArtistInfo *ArtistInfo
	Pics        []string
}

type Artists []*Artist

type Pic struct {
	PicPath    string
	ArtistInfo *ArtistInfo
}

type Pics []Pic

// ArtistsMap key: pid, value: Artist
type ArtistsMap map[string]*Artist