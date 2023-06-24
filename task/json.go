package task

import (
	"encoding/json"
	"fmt"
	"io"
)

type JsonParser[o ParseOutput] struct {
	source io.Reader
}

func NewJsonParser[o ParseOutput](source io.Reader) JsonParser[o] {
	return JsonParser[o]{
		source: source,
	}
}

func (j JsonParser[O]) Parse() (*O, error) {
	result := new(O)
	err := json.NewDecoder(j.source).Decode(result)
	if err != nil {
		return nil, fmt.Errorf("error json decoding the ouput %w", err)
	}
	return result, nil
}
