package system

// TODO - make this as interface used by:
//     * DB
//     * Cache
//     * URL Params
//     * Template Params
//     * etc.
type Stream struct {
}

/*
func (stream *Stream) Write(p []byte) (n int, err error) { // TODO: and Read()
	// remember custom was response written
	stream.responseWritten = true
	return stream.responseWriter.Write(p)
}

func (stream *Stream) GetParameter(name string, defaultValue string) string {
	value := stream.Request.FormValue(name)
	if value == "" {
		value = defaultValue
	}

	return value
}

func (stream *Stream) GetBoolParameter(name string, defaultValue bool) bool {
	value := stream.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseBool(value)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (stream *Stream) GetIntParameter(name string, defaultValue int) int {
	value := stream.Request.FormValue(name)
	if value != "" {
		result, err := strconv.Atoi(value)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (stream *Stream) GetFloatParameter(name string, defaultValue float64) float64 {
	value := stream.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (stream *Stream) GetRequiredParameter(name string) string {
	value := stream.Request.FormValue(name)
	if value == "" {
		panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
	}

	return value
}

func (stream *Stream) GetRequiredBoolParameter(name string) bool {
	value := stream.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseBool(value)
		if err == nil {
			return result
		} else {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	}

	panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
}

func (stream *Stream) GetRequiredIntParameter(name string) int {
	value := stream.Request.FormValue(name)
	if value != "" {
		result, err := strconv.Atoi(value)
		if err == nil {
			return result
		} else {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	}

	panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
}

func (stream *Stream) GetRequiredFloatParameter(name string) float64 {
	value := stream.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return result
		} else {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	}

	panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
}

func (stream *Stream) GetRequiredJSONParameter(name string, result interface{}) {
	value := stream.Request.FormValue(name)
	if value != "" {
		raw := []byte(value)
		err := json.Unmarshal(raw, result)
		if err != nil {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	} else {
		panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
	}
}
*/
