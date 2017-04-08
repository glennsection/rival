package data

import (
	"errors"
	"fmt"
	"hash/fnv"
)

// server-side data ID
type DataId uint32

// base data interface
type Data interface {
	GetDataName() string
}

// mapping from server to client IDs
var dataIdMap map[DataId]string

// convert client to server ID
func ToDataId(name string) DataId {
	h := fnv.New32a()
	h.Write([]byte(name))
	return DataId(h.Sum32())
}

// convert server to client ID
func ToDataName(id DataId) string {
	return dataIdMap[id]
}

// add ID mapping to system
func mapDataName(name string) (id DataId, err error) {
	id = ToDataId(name)
	
	if collision, ok := dataIdMap[id]; ok {
		err = errors.New(fmt.Sprintf("Collision occurred adding data to map: '%v' (collided with '%v')", name, collision))
		return
	}
	
	dataIdMap[id] = name
	return
}
