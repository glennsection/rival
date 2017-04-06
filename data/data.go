package data

import (
	"hash/fnv"
)

type DataId uint32

func ConvertToId(name string) DataId {
	h := fnv.New32a()
	h.Write([]byte(name))
	return DataId(h.Sum32())
}

func ConvertToName(id DataId) string {
	return "" // TODO
}