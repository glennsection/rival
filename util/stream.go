package util

import (
	"fmt"
	"strconv"
	"encoding/json"

	"gopkg.in/mgo.v2/bson"
)

type StreamSource interface {
	Has(name string) bool
	Set(name string, value interface{})
	Get(name string) interface{}
}

type Stream struct {
	source  StreamSource
}

func missingStreamValue(name string) string {
	return fmt.Sprintf("Missing required value: \"%v\"", name)
}

func invalidStreamValue(name string, value interface{}, err error) string {
	return fmt.Sprintf("Invalid required value: \"%v\" = \"%v\" (%v)", name, value, err)
}

func (stream *Stream) Has(name string) bool {
	return stream.source.Has(name)
}

func (stream *Stream) Set(name string, value interface{}) string {
	stream.source.Set(name, value)
	return "" // NOTE - this is to work properly with templates
}

func (stream *Stream) Get(name string) interface{} {
	return stream.source.Get(name)
}

func (stream *Stream) GetString(name string, defaultValue string) string {
	value := stream.source.Get(name)

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			return stringValue
		}
	}

	return defaultValue
}

func (stream *Stream) GetRequiredString(name string) string {
	value := stream.source.Get(name)

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			return stringValue
		}
	}

	panic(missingStreamValue(name))
}

func (stream *Stream) GetBool(name string, defaultValue bool) bool {
	value := stream.source.Get(name)

	if boolValue, ok := value.(bool); ok {
		return boolValue
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.ParseBool(stringValue)
			if err == nil {
				return result
			}
		}
	}

	return defaultValue
}

func (stream *Stream) GetRequiredBool(name string) bool {
	value := stream.source.Get(name)

	if boolValue, ok := value.(bool); ok {
		return boolValue
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.ParseBool(stringValue)
			if err == nil {
				return result
			} else {
				panic(invalidStreamValue(name, stringValue, err))
			}
		}
	}

	panic(missingStreamValue(name))
}

func (stream *Stream) GetInt(name string, defaultValue int) int {
	value := stream.source.Get(name)

	if intValue, ok := value.(int); ok {
		return intValue
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.Atoi(stringValue)
			if err == nil {
				return result
			}
		}
	}

	return defaultValue
}

func (stream *Stream) GetRequiredInt(name string) int {
	value := stream.source.Get(name)

	if intValue, ok := value.(int); ok {
		return intValue
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.Atoi(stringValue)
			if err == nil {
				return result
			} else {
				panic(invalidStreamValue(name, stringValue, err))
			}
		}
	}

	panic(missingStreamValue(name))
}

func (stream *Stream) GetFloat(name string, defaultValue float64) float64 {
	value := stream.source.Get(name)

	if floatValue, ok := value.(float64); ok {
		return floatValue
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.ParseFloat(stringValue, 64)
			if err == nil {
				return result
			}
		}
	}

	return defaultValue
}

func (stream *Stream) GetRequiredFloat(name string) float64 {
	value := stream.source.Get(name)

	if floatValue, ok := value.(float64); ok {
		return floatValue
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.ParseFloat(stringValue, 64)
			if err == nil {
				return result
			} else {
				panic(invalidStreamValue(name, stringValue, err))
			}
		}
	}

	panic(missingStreamValue(name))
}

func (stream *Stream) GetJSON(name string, result interface{}) {
	value := stream.source.Get(name)

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			raw := []byte(stringValue)
			err := json.Unmarshal(raw, result)
			if err == nil {
				return
			}
		}
	}
	return
}

func (stream *Stream) GetRequiredJSON(name string, result interface{}) {
	value := stream.source.Get(name)

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			raw := []byte(stringValue)
			err := json.Unmarshal(raw, result)
			if err == nil {
				return
			} else {
				panic(invalidStreamValue(name, stringValue, err))
			}
		}
	}

	panic(missingStreamValue(name))
}

func (stream *Stream) GetID(name string) bson.ObjectId {
	value := stream.source.Get(name)

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			return bson.ObjectIdHex(stringValue)
		}
	}

	return bson.ObjectId("")
}

func (stream *Stream) GetRequiredID(name string) bson.ObjectId {
	value := stream.source.Get(name)

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			return bson.ObjectIdHex(stringValue)
		}
	}

	panic(missingStreamValue(name))
}
