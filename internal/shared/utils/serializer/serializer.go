package serializer

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// init registers certain types we will be using in the serializer
//
// *NOTE: This function gets run automatically when the package is imported
func init() {
	performRegistrations()
}

func performRegistrations() {
	log.Println("Registering types with gob...")
	gob.Register(time.Time{})
	gob.Register(map[string]any{})
	gob.Register(map[string]string{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[string]bool{})
}

func SerializeMap(m *map[string]any) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DeserializeMap(data []byte) (*map[string]any, error) {
	var m map[string]any
	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
