package data

import (
	"bloodtales/util"
)

type QuestObjectivesStreamSource struct {
	bindings		map[string]interface{}	`bson:"bi"`	
}

func NewQuestObjectivesStreamSource() *util.Stream {
	stream := util.Stream {}
	source := QuestObjectivesStreamSource {
		bindings: map[string]interface{} {},
	}
	stream.SetSource(source)
	return &stream
}

func (source QuestObjectivesStreamSource) Has(name string) bool {
	_, hasName := source.bindings[name]
	return hasName
}

func (source QuestObjectivesStreamSource) Set(name string, value interface{}) {
	source.bindings[name] = value
}

func (source QuestObjectivesStreamSource) Get(name string) interface{} {
	val, hasName := source.bindings[name]

	if hasName {
		return val
	}

	return nil
}