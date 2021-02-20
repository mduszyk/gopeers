package udprpc

import (
	"bytes"
	"errors"
	"log"
	"testing"
)

func TestJsonEncoder(t *testing.T) {
	message := RpcMessage {
		Type: RpcTypeRequest,
		Service: RpcService(1),
		Id: RpcId(1),
		Payload: []byte{2, 4, 8, 16},
		Error: []byte(errors.New("test error").Error()),
	}
	encoder := NewJsonEncoder()
	buf, err := encoder.Encode(message)
	if err != nil {
		t.Errorf("failed encoding: %v\n", err)
	}
	log.Printf("encoded message: %s\n", buf)
	var decodedMessage RpcMessage
	decoder := NewJsonDecoder()
	err = decoder.Decode(buf, &decodedMessage)
	if err != nil {
		t.Errorf("failed decoding: %v\n", err)
	}
	if decodedMessage.Id != message.Id ||
		decodedMessage.Type != message.Type ||
		decodedMessage.Service != message.Service ||
		!bytes.Equal(decodedMessage.Payload, message.Payload) ||
		!bytes.Equal(decodedMessage.Error, message.Error) {
		t.Errorf("message decoded incorrectly\n")
	}
	log.Printf("decoded message: %+v\n", decodedMessage)
}
