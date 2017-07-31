package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const ReplayInfoCollectionName = "replayInfos"

type ReplayInfo struct {
	ID     bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserID bson.ObjectId `bson:"us" json:"-"`
	Info   string        `bson:"in" json:"info"`
}

const ReplayDataCollectionName = "replayDatas"

type ReplayData struct {
	ID     bson.ObjectId `bson:"_id,omitempty" json:"-"`
	InfoID bson.ObjectId `bson:"iid" json:"-"`
	Data   string        `bson:"dt" json:"data"`
}

func ensureIndexReplay(database *mgo.Database) {
	c := database.C(ReplayInfoCollectionName)

	// user index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string{"us"},
		Background: true,
	}))

	c = database.C(ReplayDataCollectionName)

	// user index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string{"iid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
	}))
}

func CreateReplay(context *util.Context, info string, data string) (err error) {
	// init replay info
	replayInfo := &ReplayInfo{
		UserID: context.UserID,
		Info:   info,
	}

	// save replay info
	err = replayInfo.Save(context)
	if err != nil {
		return
	}

	// init replay data
	replayData := &ReplayData{
		InfoID: replayInfo.ID,
		Data:   data,
	}

	// save replay data
	err = replayData.Save(context)
	return
}

func GetReplayInfosByUser(context *util.Context, userId bson.ObjectId) (replayInfos []*ReplayInfo, err error) {
	// find replay infos by user ID
	err = context.DB.C(ReplayInfoCollectionName).Find(bson.M{"us": userId}).All(&replayInfos)
	return
}

func GetReplayInfoById(context *util.Context, infoId bson.ObjectId) (replayInfo *ReplayInfo, err error) {
	err = context.DB.C(ReplayInfoCollectionName).Find(bson.M{"_id": infoId}).One(&replayInfo)
	return
}

func GetReplayDataByInfo(context *util.Context, infoId bson.ObjectId) (replayData *ReplayData, err error) {
	// find replay data by info ID
	err = context.DB.C(ReplayDataCollectionName).Find(bson.M{"iid": infoId}).One(&replayData)
	return
}

func (replayInfo *ReplayInfo) Save(context *util.Context) (err error) {
	replayInfo.ID = bson.NewObjectId()

	err = context.DB.C(ReplayInfoCollectionName).Insert(replayInfo)
	return
}

func (replayData *ReplayData) Save(context *util.Context) (err error) {
	replayData.ID = bson.NewObjectId()

	err = context.DB.C(ReplayDataCollectionName).Insert(replayData)
	return
}

func (replayInfo *ReplayInfo) Delete(context *util.Context) (err error) {
	// delete replay data from database
	err = context.DB.C(ReplayDataCollectionName).Remove(bson.M{"iid": replayInfo.ID})
	if err != nil {
		return
	}

	// delete replay info from database
	return context.DB.C(ReplayInfoCollectionName).Remove(bson.M{"_id": replayInfo.ID})
}
