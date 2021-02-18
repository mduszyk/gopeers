package udprpc

import "encoding/json"

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

type Decoder interface {
	Decode(p []byte, v interface{}) error
}

type jsonEncoder struct { }

func NewJsonEncoder() Encoder {
	return &jsonEncoder{}
}

func (encoder *jsonEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

type jsonDecoder struct { }

func NewJsonDecoder() Decoder {
	return &jsonDecoder{}
}

func (encoder *jsonDecoder) Decode(p []byte, v interface{}) error {
	return json.Unmarshal(p, v)
}
