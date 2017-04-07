package data

import (
	"errors"
	"fmt"
	"hash/fnv"
	"log"
)

type DataId uint32

type Data interface {
	GetDataName() string
}

type DataMap struct {
	Table map[DataId]Data
}

var dataMap *DataMap = nil

func GetDataId(name string) DataId {
	h := fnv.New32a()
	h.Write([]byte(name))
	return DataId(h.Sum32())
}

func Load() {
	dataMap = &DataMap {
		Table: map[DataId]Data {},
	}
	
	// HACK
	AddData(CardData { Name: "CARD_1" })
	AddData(CardData { Name: "CARD_2" })
	AddData(CardData { Name: "CARD_3" })
}

func AddData(data Data) (err error) {
	name := data.GetDataName()
	id := GetDataId(name)
	
	if collision, ok := dataMap.Table[id]; ok {
		collisionName := collision.GetDataName()
		err = errors.New(fmt.Sprintf("Collision occurred adding data to map: %v (collided with %v)", name, collisionName))
		return
	}
	
	log.Printf("Adding data: %v", id) // DEBUG
	dataMap.Table[id] = data
	return
}

func GetDataById(id DataId) Data {
	return dataMap.Table[id]
}

func GetDataByName(name string) Data {
	id := GetDataId(name)
	return GetDataById(id)
}
