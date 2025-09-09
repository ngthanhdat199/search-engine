package pkg

import "encoding/json"

func ToJSON(input interface{}) string {
	jsonRaw, err := json.Marshal(input)
	if err != nil {
		return "failed to encoding instance"
	}

	return string(jsonRaw)
}
