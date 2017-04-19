package data

import (
	"fmt"
	"strings"
	"math"
	"encoding/json"
)

type RankData struct {
	Level                   int           `json:"rank,string"`
	StartRank               int           `json:"startingStars,string"`
	NextRank                int           `json:"starsForNextLevel,string"`
	Image                   string 	      `json:"imagePrefab"`
}

// data map
var ranks []*RankData

// implement Data interface
func (data *RankData) GetDataName() string {
	return fmt.Sprintf("Rank %d", data.Level)
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type RanksParsed struct {
	PvPRanking []RankData
}

// data processor
func LoadRanks(raw []byte) {
	// parse
	container := &RanksParsed {}
	json.Unmarshal(raw, container)

	// enter into system data
	for i, _ := range container.PvPRanking {
		// insert into table
		ranks = append(ranks, &container.PvPRanking[i])
	}
}

// get rank by level
func GetRankForLevel(level int) *RankData {
	return ranks[level - 1]
}

// get rank by points
func GetRank(rank int) *RankData {
	for _, rankData := range ranks {
		if rank < rankData.NextRank || rankData.NextRank < 0 {
			return rankData
		}
	}
	return nil
}

func (rank *RankData) GetTier() int {
	return int(math.Ceil(float64(rank.Level) / 5.0))
}

func (rank *RankData) GetImageSrc() string {
	src := rank.Image
	idx := strings.LastIndex(src, "/")
	if idx >= 0 {
		src = src[idx + 1:]
	}
	return fmt.Sprintf("/static/img/ranks/%v.png", src)
}