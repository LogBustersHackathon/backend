package utils

import "encoding/json"

func ConvertToJSON[T any](object T) ([]byte, error) {
	jsonOut, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	return jsonOut, nil
}

func ConvertFromJSON[T any](data []byte) (T, error) {
	var object T

	if len(data) == 0 {
		return object, nil
	}

	err := json.Unmarshal(data, &object)
	if err != nil {
		return object, err
	}

	return object, nil
}
