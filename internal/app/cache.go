package app

import (
	"bytes"
	"encoding/json"
	"errors"
)

type CacheType int

const (
	CacheLocal CacheType = iota + 1
	CacheRedis
)

var cacheTypeToString = map[CacheType]string{
	CacheLocal: "local",
	CacheRedis: "redis",
}

var cacheTypeToID = map[string]CacheType{
	"local": CacheLocal,
	"redis": CacheRedis,
}

func (st CacheType) String() string {
	return cacheTypeToString[st]
}

// MarshalJSON marshals the enum as a quoted json string
func (s CacheType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(cacheTypeToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *CacheType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*s = cacheTypeToID[j]
	return nil
}

// MarshalYAML marshals the enum as a YAML string
func (s CacheType) MarshalYAML() (interface{}, error) {
	return cacheTypeToString[s], nil
}

// UnmarshalYAML unmarshals a YAML string to the enum value
func (s *CacheType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var y string
	if err := unmarshal(&y); err != nil {
		return err
	}
	*s = cacheTypeToID[y]
	return nil
}

func (e *CacheType) UnmarshalTOML(data interface{}) error {
	if str, ok := data.(string); ok {
		if v, ok := cacheTypeToID[str]; ok {
			*e = v
			return nil
		}
		return errors.New("invalid CacheType")
	}
	return errors.New("expected string for CacheType")
}

func (e CacheType) MarshalTOML() (interface{}, error) {
	return cacheTypeToString[e], nil
}
