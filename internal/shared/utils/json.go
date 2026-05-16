package utils

import (
	"encoding/json"
	"fmt"
)

func ParseJSON(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return nil
}

func ToIndentedJson(v any) string {
	json, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		println("ToIndentedJson error: ", err.Error())
		return ""
	}
	return string(json)
}

func MapStruct(dest any, src any) {
	data, _ := json.Marshal(src)
	_ = json.Unmarshal(data, dest)
}
