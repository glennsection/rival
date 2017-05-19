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
func ToDataName(id DataId) (name string) {
	name = dataIdMap[id]
	if name == "" {
		name = "INVALID"
	}
	return
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

//implementing sort.Interface for use with sort.Sort function
type DataIdCollection []DataId

func (arr DataIdCollection) Len() int {
	return len(arr)
}

func (arr DataIdCollection) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func (arr DataIdCollection) Less(i, j int) bool {
	return uint32(arr[i]) < uint32(arr[j])
}
