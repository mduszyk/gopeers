package udprpc

import (
	"bytes"
	"log"
	"testing"
)

var echo1Payloads []RpcPayload
var echo2Payloads []RpcPayload
var echo3Payloads []RpcPayload

func echo1(payload RpcPayload) RpcPayload {
	log.Printf("echo1 received payload: %s\n", payload)
	echo1Payloads = append(echo1Payloads, payload)
	return payload
}

func echo2(payload RpcPayload) RpcPayload {
	log.Printf("echo2 received payload: %s\n", payload)
	echo2Payloads = append(echo2Payloads, payload)
	return payload
}

func echo3(payload RpcPayload) RpcPayload {
	log.Printf("echo3 received payload: %s\n", payload)
	echo3Payloads = append(echo3Payloads, payload)
	return payload
}

func TestRpcNode(t *testing.T) {
	node1, err := NewRpcNode("localhost:", []RpcFunc{echo1, echo2})
	if err != nil {
		t.Errorf("failed creating rpc node: %v\n", err)
	}
	go node1.Run()
	node2, err := NewRpcNode("localhost:", []RpcFunc{echo3})
	if err != nil {
		t.Errorf("failed creating rpc node: %v\n", err)
	}
	go node2.Run()

	response, err := node2.Call(node1.Addr, RpcService(0), []byte("test1"))
	if err != nil {
		t.Errorf("failed calling rpc service: %v\n", err)
	}
	if !bytes.Equal(response, []byte("test1")) {
		t.Errorf("rpc service returned invalid response: %s\n", response)
	}
	if len(echo1Payloads) != 1 {
		t.Errorf("rpc service was not called the correct number of times\n")
	}
	if !bytes.Equal(echo1Payloads[0], []byte("test1")) {
		t.Errorf("rpc service was not called\n")
	}

	response, err = node2.Call(node1.Addr, RpcService(1), []byte("test2"))
	if err != nil {
		t.Errorf("failed calling rpc service: %v\n", err)
	}
	if !bytes.Equal(response, []byte("test2")) {
		t.Errorf("rpc service returned invalid response: %s\n", response)
	}
	if len(echo2Payloads) != 1 {
		t.Errorf("rpc service was not called the correct number of times\n")
	}
	if !bytes.Equal(echo2Payloads[0], []byte("test2")) {
		t.Errorf("rpc service was not called\n")
	}

	response, err = node1.Call(node2.Addr, RpcService(0), []byte("test3"))
	if err != nil {
		t.Errorf("failed calling rpc service: %v\n", err)
	}
	if !bytes.Equal(response, []byte("test3")) {
		t.Errorf("rpc service returned invalid response: %s\n", response)
	}
	if len(echo3Payloads) != 1 {
		t.Errorf("rpc service was not called the correct number of times\n")
	}
	if !bytes.Equal(echo3Payloads[0], []byte("test3")) {
		t.Errorf("rpc service was not called\n")
	}
}
