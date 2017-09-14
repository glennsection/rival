package util

import (
	"fmt"
	"strings"
	"strconv"
	"errors"
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

func (stream *Stream) SetSource(newSource StreamSource) bool {
	if stream.source != nil {
		return false
	}

	stream.source = newSource
	return true
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

func (stream *Stream) GetRequiredStrings(name string) []string {
	value := stream.GetRequiredString(name)

	subvalues := strings.Split(value, ",")

	results := make([]string, len(subvalues))
	for i, subvalue := range subvalues {
		results[i] = subvalue
	}

	return results
}

func parseBool(name string, value interface{}) (bool, error) {
	if boolValue, ok := value.(bool); ok {
		return boolValue, nil
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.ParseBool(stringValue)
			if err == nil {
				return result, nil
			} else {
				return false, errors.New(invalidStreamValue(name, value, err))
			}
		}
	}

	return false, errors.New(missingStreamValue(name))
}

func (stream *Stream) GetBool(name string, defaultValue bool) bool {
	value := stream.source.Get(name)

	result, err := parseBool(name, value)
	if err != nil {
		return defaultValue
	}

	return result
}

func (stream *Stream) GetRequiredBool(name string) bool {
	value := stream.source.Get(name)

	result, err := parseBool(name, value)
	if err != nil {
		panic(err)
	}

	return result
}

func (stream *Stream) GetRequiredBools(name string) []bool {
	value := stream.GetRequiredString(name)

	subvalues := strings.Split(value, ",")

	results := make([]bool, len(subvalues))
	var err error
	for i, subvalue := range subvalues {
		results[i], err = parseBool(fmt.Sprintf("%s[%d]", name, i), subvalue)
		if err != nil {
			panic(err)
		}
	}

	return results
}

func parseInt(name string, value interface{}) (int, error) {
	if intValue, ok := value.(int); ok {
		return intValue, nil
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.Atoi(stringValue)
			if err == nil {
				return result, nil
			} else {
				return 0, errors.New(invalidStreamValue(name, value, err))
			}
		}
	}

	return 0, errors.New(missingStreamValue(name))
}

func (stream *Stream) GetInt(name string, defaultValue int) int {
	value := stream.source.Get(name)

	result, err := parseInt(name, value)
	if err != nil {
		return defaultValue
	}

	return result
}

func (stream *Stream) GetRequiredInt(name string) int {
	value := stream.source.Get(name)

	result, err := parseInt(name, value)
	if err != nil {
		panic(err)
	}

	return result
}

func (stream *Stream) GetRequiredInts(name string) []int {
	value := stream.GetRequiredString(name)

	subvalues := strings.Split(value, ",")

	results := make([]int, len(subvalues))
	var err error
	for i, subvalue := range subvalues {
		results[i], err = parseInt(fmt.Sprintf("%s[%d]", name, i), subvalue)
		if err != nil {
			panic(err)
		}
	}

	return results
}

func parseInt64(name string, value interface{}) (int64, error) {
	if int64Value, ok := value.(int64); ok {
		return int64Value, nil
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.ParseInt(stringValue, 10, 64)
			if err == nil {
				return result, nil
			} else {
				return 0, errors.New(invalidStreamValue(name, value, err))
			}
		}
	}

	return 0, errors.New(missingStreamValue(name))
}

func (stream *Stream) GetInt64(name string, defaultValue int64) int64 {
	value := stream.source.Get(name)

	result, err := parseInt64(name, value)
	if err != nil {
		return defaultValue
	}

	return result
} 

func (stream *Stream) GetRequiredInt64(name string) int64 {
	value := stream.source.Get(name)

	result, err := parseInt64(name, value)
	if err != nil {
		panic(err)
	}

	return result
}

func (stream *Stream) GetRequiredInt64s(name string) []int64 {
	value := stream.GetRequiredString(name)

	subvalues := strings.Split(value, ",")

	results := make([]int64, len(subvalues))
	var err error
	for i, subvalue := range subvalues {
		results[i], err = parseInt64(fmt.Sprintf("%s[%d]", name, i), subvalue)
		if err != nil {
			panic(err)
		}
	}

	return results
}

func parseFloat(name string, value interface{}) (float64, error) {
	if floatValue, ok := value.(float64); ok {
		return floatValue, nil
	}

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			result, err := strconv.ParseFloat(stringValue, 64)
			if err == nil {
				return result, nil
			} else {
				return 0, errors.New(invalidStreamValue(name, value, err))
			}
		}
	}

	return 0, errors.New(missingStreamValue(name))
}

func (stream *Stream) GetFloat(name string, defaultValue float64) float64 {
	value := stream.source.Get(name)

	result, err := parseFloat(name, value)
	if err != nil {
		return defaultValue
	}

	return result
}

func (stream *Stream) GetRequiredFloat(name string) float64 {
	value := stream.source.Get(name)

	result, err := parseFloat(name, value)
	if err != nil {
		panic(err)
	}

	return result
}

func (stream *Stream) GetRequiredFloats(name string) []float64 {
	value := stream.GetRequiredString(name)

	subvalues := strings.Split(value, ",")

	results := make([]float64, len(subvalues))
	var err error
	for i, subvalue := range subvalues {
		results[i], err = parseFloat(fmt.Sprintf("%s[%d]", name, i), subvalue)
		if err != nil {
			panic(err)
		}
	}

	return results
}

func (stream *Stream) GetJSON(name string, result interface{}) bool {
	value := stream.source.Get(name)

	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			raw := []byte(stringValue)
			err := json.Unmarshal(raw, result)
			if err == nil {
				return true
			} else {
				panic(invalidStreamValue(name, stringValue, err))
			}
		}
	}
	return false
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

func parseId(name string, value interface{}) (bson.ObjectId, error) {
	if stringValue, ok := value.(string); ok {
		if stringValue != "" {
			if bson.IsObjectIdHex(stringValue) {
				return bson.ObjectIdHex(stringValue), nil
			}
		}
	}

	return bson.ObjectId(""), errors.New(missingStreamValue(name))
}

func (stream *Stream) GetId(name string) bson.ObjectId {
	value := stream.source.Get(name)

	result, err := parseId(name, value)
	if err != nil {
		return bson.ObjectId("")
	}

	return result
}

func (stream *Stream) GetRequiredId(name string) bson.ObjectId {
	value := stream.source.Get(name)

	result, err := parseId(name, value)
	if err != nil {
		panic(err)
	}

	return result
}

func (stream *Stream) GetRequiredIds(name string) []bson.ObjectId {
	value := stream.GetRequiredString(name)

	subvalues := strings.Split(value, ",")

	results := make([]bson.ObjectId, len(subvalues))
	var err error
	for i, subvalue := range subvalues {
		results[i], err = parseId(fmt.Sprintf("%s[%d]", name, i), subvalue)
		if err != nil {
			panic(err)
		}
	}

	return results
}