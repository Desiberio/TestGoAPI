package main

import (
	"bytes"
	"encoding/json"
)

type News struct {
	Id         int64   `json:"Id" db:"Id"`
	Title      string  `json:"Title" db:"Title"`
	Content    string  `json:"Content" db:"Content"`
	Categories []int64 `json:"Categories" db:"Categories"`
}

func (t *News) JSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
