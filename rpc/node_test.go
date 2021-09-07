package rpc

import (
	"bytes"
	"errors"
	"log"
	"net"
	"testing"
	"time"
)

var echo1Payloads []Payload
var echo2Payloads []Payload
var echo3Payloads []Payload
var failurePayloads []Payload

var callTimeout = time.Second
var bufferSize = uint32(1024)

func echo1(addr *net.UDPAddr, payload Payload) (Payload, error) {
	log.Printf("echo1 received payload: %s\n", payload)
	echo1Payloads = append(echo1Payloads, payload)
	return payload, nil
}

func echo2(addr *net.UDPAddr, payload Payload) (Payload, error) {
	log.Printf("echo2 received payload: %s\n", payload)
	echo2Payloads = append(echo2Payloads, payload)
	return payload, nil
}

func echo3(addr *net.UDPAddr, payload Payload) (Payload, error) {
	log.Printf("echo3 received payload: %s\n", payload)
	echo3Payloads = append(echo3Payloads, payload)
	return payload, nil
}

func failure(addr *net.UDPAddr, payload Payload) (Payload, error) {
	log.Printf("failure received payload: %s\n", payload)
	failurePayloads = append(failurePayloads, payload)
	return nil, errors.New("rpc service failure")
}

func TestRpcNode(t *testing.T) {
	node1, err := NewUdpNode("localhost:", []Service{echo1, echo2}, callTimeout, bufferSize)
	if err != nil {
		t.Errorf("failed creating rpc node: %v\n", err)
	}
	go node1.Run()
	node2, err := NewUdpNode("localhost:", []Service{echo3, failure}, callTimeout, bufferSize)
	if err != nil {
		t.Errorf("failed creating rpc node: %v\n", err)
	}
	go node2.Run()

	response, err := node2.Call(node1.Addr, ServiceId(0), []byte("test1"))
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

	response, err = node2.Call(node1.Addr, ServiceId(1), []byte("test2"))
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

	response, err = node1.Call(node2.Addr, ServiceId(0), []byte("test3"))
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

	response, err = node1.Call(node2.Addr, ServiceId(1), []byte("test4"))
	if err == nil || err.Error() != "rpc service failure" {
		t.Errorf("expected error from rpc service\n")
	}
	if len(failurePayloads) != 1 {
		t.Errorf("rpc service was not called the correct number of times\n")
	}
	if !bytes.Equal(failurePayloads[0], []byte("test4")) {
		t.Errorf("rpc service was not called\n")
	}
}
