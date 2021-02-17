package udprpc

import (
	"bytes"
	"log"
	"testing"
)

func TestJsonEncoder(t *testing.T) {
	id := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	payload := []byte{2, 4, 8, 16}
	message := RpcMessage {
		Id: id,
		Type: TypeService1Request,
		Payload: payload,
	}
	encoder := &JsonEncoder{}
	buf, err := encoder.Encode(message)
	if err != nil {
		t.Errorf("failed encoding: %v\n", err)
	}
	log.Printf("encoded message: %s\n", buf)
	var decodedMessage RpcMessage
	err = encoder.Decode(buf, &decodedMessage)
	if err != nil {
		t.Errorf("failed decoding: %v\n", err)
	}
	if decodedMessage.Id != message.Id ||
		decodedMessage.Type != message.Type ||
		!bytes.Equal(decodedMessage.Payload, message.Payload) {
		t.Errorf("message decoded incorrectly\n")
	}
	log.Printf("decoded message: %+v\n", decodedMessage)
}
