package models

import (
	"bloodtales/util"
	"bloodtales/data"
)

const (
	Tutorial_NewPlayer util.Bits = 0x0
	Tutorial_TomeClaimed = 0x1
	Tutorial_TomeOpened = 0x2
)

func (player *Player) ClaimTutorialTome(context *util.Context) {
	if (player.TutorialProgress & Tutorial_TomeClaimed) == Tutorial_TomeClaimed {
		return
	}

	_, tome := player.GetEmptyTomeSlot()
	if tome != nil {
		tome.DataID = data.ToDataId("TOME_TUTORIAL")
		tome.State = TomeLocked
		tome.UnlockTime = 0
	}

	player.TutorialProgress |= Tutorial_TomeClaimed
	player.Save(context)
}

func (player *Player) OpenTutorialTome(context *util.Context, tome *Tome) *Reward {
	if (player.TutorialProgress & Tutorial_TomeOpened) == Tutorial_TomeOpened || tome == nil {
		return nil
	}

	tome.DataID = 0
	tome.State = TomeEmpty
	tome.UnlockTime = 0

	reward := &Reward {
		Cards: []data.DataId {data.ToDataId("CARD_ARMORED_SKELETON"), data.ToDataId("CARD_SKULL_SWARM"), data.ToDataId("CARD_FARMER"), data.ToDataId("CARD_LIGHTNING_SWORD")},
		NumRewarded: []int {15, 5, 5, 2},
		PremiumCurrency: 2,
		StandardCurrency: 25,
	}

	player.AddRewards(reward, context)

	player.TutorialProgress |= Tutorial_TomeOpened
	player.Save(context)

	return reward
}