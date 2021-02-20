package udprpc

import (
	"bytes"
	"log"
	"testing"
)

func echo(payload RpcPayload) RpcPayload {
	log.Printf("echo received payload: %s\n", payload)
	return payload
}

// {"Type":1,"Service":0,"Id":5577006791947779410,"Payload":"dGVzdA=="}
func TestRpcNode(t *testing.T) {
	node1, err := NewRpcNode("localhost:7777", []RpcFunc{echo})
	if err != nil {
		t.Errorf("failed creating rpc node: %v\n", err)
	}
	go node1.Run()
	var services2 []RpcFunc
	node2, err := NewRpcNode("localhost:", services2)
	if err != nil {
		t.Errorf("failed creating rpc node: %v\n", err)
	}
	go node2.Run()

	response, err := node2.Call(node1.Addr, RpcService(0), []byte("test"))
	if err != nil {
		t.Errorf("failed calling rpc service: %v\n", err)
	}
	if !bytes.Equal(response, []byte("test")) {
		t.Errorf("rpc service returned invalid response: %s\n", response)
	}
}
