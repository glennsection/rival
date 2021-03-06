package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const ReplayInfoCollectionName = "replayInfos"

type ReplayInfo struct {
	ID         bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserID     bson.ObjectId `bson:"us" json:"-"`
	CreatedAt  time.Time     `bson:"t0" json:"created"`
	Rank       int           `bson:"rk" json:"rank"`
	Info       string        `bson:"in" json:"info"`
}

const ReplayDataCollectionName = "replayDatas"

type ReplayData struct {
	ID         bson.ObjectId `bson:"_id,omitempty" json:"-"`
	InfoID     bson.ObjectId `bson:"iid" json:"-"`
	Data       string        `bson:"dt" json:"data"`
}

func ensureIndexReplay(database *mgo.Database) {
	c := database.C(ReplayInfoCollectionName)

	// user index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string { "us" },
		Background: true,
	}))

	// user/time index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string { "us", "t0" },
		Background: true,
	}))

	c = database.C(ReplayDataCollectionName)

	// user index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string { "iid" },
		Unique:     true,
		DropDups:   true,
		Background: true,
	}))
}

func CreateReplay(context *util.Context, info string, data string, rank int) (err error) {
	// init replay info
	replayInfo := &ReplayInfo {
		UserID: context.UserID,
		Info:   info,
		Rank:   rank,
	}

	// save replay info
	err = replayInfo.Save(context)
	if err != nil {
		return
	}

	// init replay data
	replayData := &ReplayData {
		InfoID: replayInfo.ID,
		Data:   data,
	}

	// save replay data
	err = replayData.Save(context)
	return
}

func GetTopReplays(context *util.Context, userId bson.ObjectId) (replayInfos []*ReplayInfo, err error) {
	// find replay infos by user ID
	err = context.DB.C(ReplayInfoCollectionName).Find(nil).Sort("-rk").Limit(5).All(&replayInfos)
	return
}

func GetReplayInfosByUser(context *util.Context, userId bson.ObjectId) (replayInfos []*ReplayInfo, err error) {
	// find replay infos by user ID
	//err = context.DB.C(ReplayInfoCollectionName).Find(bson.M { "us": userId }).All(&replayInfos)
	err = context.DB.C(ReplayInfoCollectionName).Find(bson.M { "us": userId }).Sort("-t0").Limit(10).All(&replayInfos)
	return
}

func GetAllReplayInfosByUser(context *util.Context, userId bson.ObjectId) (replayInfos []*ReplayInfo, err error) {
	// find replay infos by user ID
	//err = context.DB.C(ReplayInfoCollectionName).Find(bson.M { "us": userId }).All(&replayInfos)
	err = context.DB.C(ReplayInfoCollectionName).Find(bson.M { "us": userId }).Sort("-t0").All(&replayInfos)
	return
}

func GetLastReplayInfoByUser(context *util.Context, userId bson.ObjectId) (replayInfo *ReplayInfo, err error) {
	// find last replay info by user ID
	err = context.DB.C(ReplayInfoCollectionName).Find(bson.M { "us": userId }).Limit(1).Sort("-t0").One(&replayInfo)
	return
}

func GetReplayInfoById(context *util.Context, infoId bson.ObjectId) (replayInfo *ReplayInfo, err error) {
	err = context.DB.C(ReplayInfoCollectionName).Find(bson.M { "_id": infoId }).One(&replayInfo)
	return
}

func GetReplayDataByInfo(context *util.Context, infoId bson.ObjectId) (replayData *ReplayData, err error) {
	// find replay data by info ID
	err = context.DB.C(ReplayDataCollectionName).Find(bson.M { "iid": infoId }).One(&replayData)
	return
}

func (replayInfo *ReplayInfo) Save(context *util.Context) (err error) {
	replayInfo.ID = bson.NewObjectId()
	replayInfo.CreatedAt = time.Now()

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
	err = context.DB.C(ReplayDataCollectionName).Remove(bson.M { "iid": replayInfo.ID })
	if err == mgo.ErrNotFound {
		err = nil
	}
	if err != nil {
		return
	}

	// delete replay info from database
	return context.DB.C(ReplayInfoCollectionName).Remove(bson.M{"_id": replayInfo.ID})
}
