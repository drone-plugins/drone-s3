package main

import (
	"encoding/json"
)

type StringMapFlag struct {
	parts map[string]string
}

func (s *StringMapFlag) String() string {
	return ""
}

func (s *StringMapFlag) Get() map[string]string {
	return s.parts
}

func (s *StringMapFlag) Set(value string) error {
	s.parts = map[string]string{}
	err := json.Unmarshal([]byte(value), &s.parts)
	if err != nil {
		s.parts["*"] = value
	}
	return nil
}
