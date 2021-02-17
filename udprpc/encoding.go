package udprpc

import "encoding/json"

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

type Decoder interface {
	Decode(p []byte, v interface{}) error
}

type JsonEncoder struct { }

func (encoder *JsonEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (encoder *JsonEncoder) Decode(p []byte, v interface{}) error {
	return json.Unmarshal(p, v)
}
