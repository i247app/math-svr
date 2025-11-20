package requtil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type WrappedRequest struct {
	requestBodyJson map[string]any
}

func Wrap(req *http.Request) (*WrappedRequest, error) {
	rawBody, err := safelyCopyRequestBody(req)
	if err != nil {
		return nil, err
	}

	var mapBody map[string]any
	if err := json.Unmarshal(rawBody.Bytes(), &mapBody); err != nil {
		return nil, err
	}

	// Reset the request body so it can be read again down the line
	req.Body = io.NopCloser(rawBody)

	return &WrappedRequest{mapBody}, nil
}

func safelyCopyRequestBody(req *http.Request) (*bytes.Buffer, error) {
	rawBody := new(bytes.Buffer)
	if _, err := rawBody.ReadFrom(req.Body); err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(rawBody)
	return rawBody, nil
}

type RequestMetadata struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

func (w *WrappedRequest) Metadata() (*RequestMetadata, error) {
	metadata, ok := w.requestBodyJson["__metadata"].(map[string]any)
	if !ok {
		return nil, errors.New("metadata not found or invalid format")
	}

	deviceID, _ := metadata["device_id"].(string)
	deviceName, _ := metadata["device_name"].(string)

	res := RequestMetadata{
		DeviceID:   deviceID,
		DeviceName: deviceName,
	}
	return &res, nil
}
