package data

import(
	"encoding/json"

	"bloodtales/util"
)

type League int
const (
	_ 		 				League 		= iota 
	WoodLeague 							//leagues start at 1
	BronzeLeague
	SilverLeague
	GoldLeague
	PlatinumLeague
	ChampionsLeague
)

type LeagueData struct {
	ID 						string 		`json:"id"`
	RankTier 				int 		`json:"rankTier,string"`
	RankMin 				int 		`json:"rankMin,string"`
	RankMax 				int 		`json:"rankMax,string"`
	TomeVolumeMultiplier 	float64 	`json:"tomeVolumeMultiplier,string"`
	TomePriceMultiplier 	float64 	`json:"tomePriceMultiplier,string"`
}

var leagues map[League]*LeagueData

type LeagueDataParsed struct {
	PvPLeagues 				[]LeagueData
}

func LoadLeagues(raw []byte) {
	container := &LeagueDataParsed{}
	util.Must(json.Unmarshal(raw, container))

	leagues = map[League]*LeagueData{}
	for _, leagueData := range container.PvPLeagues {
		league := League(leagueData.RankTier)

		leagues[league] = &leagueData
	}
}

func GetLeagueData(val League) *LeagueData {
	return leagues[val]
}

func GetLeague(rank int) League {
	for _, data := range leagues {
		if rank >= data.RankMin && rank <= data.RankMax {
			return League(data.RankTier)
		}
	}

	return WoodLeague
}

func GetLeagueByID(id string) League {
	for _, data := range leagues {
		if id == data.ID {
			return League(data.RankTier)
		}
	}

	return WoodLeague
}