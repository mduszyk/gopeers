package peers

import (
	"testing"
	"time"
)

func TestUdpProtocol(t *testing.T) {
	node1Id, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random id: %v\n", err)
	}
	node1Peer := Peer{Id: node1Id, LastSeen: time.Now()}
	node1BucketList := NewBucketList(20, 5, node1Peer)
	p2pNode1 := NewP2pNode(node1Id, node1BucketList)
	protoServer1, err := NewUdpProtocolServer("localhost:5001", p2pNode1)
	node1Peer.Protocol = protoServer1.Connect(protoServer1.rpcNode.Addr)

	node2Id, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random id: %v\n", err)
	}
	node2Peer := Peer{Id: node2Id, LastSeen: time.Now()}
	node2BucketList := NewBucketList(20, 5, node2Peer)
	p2pNode2 := NewP2pNode(node2Id, node2BucketList)
	protoServer2, err := NewUdpProtocolServer("localhost:5002", p2pNode2)
	node2Peer.Protocol = protoServer2.Connect(protoServer2.rpcNode.Addr)

	node1Protocol := protoServer2.Connect(protoServer1.rpcNode.Addr)
	//node2Protocol := protoServer1.Connect(protoServer2.rpcNode.Addr)

	randomId, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id: %v\n", err)
	}
	echoId, err := node1Protocol.Ping(node2Peer, randomId)
	if err != nil {
		t.Errorf("failed pinging: %v\n", err)
	}
	if echoId.Cmp(randomId) != 0 {
		t.Errorf("ping returned invalid Id\n")
	}
	if i, _ := p2pNode1.buckets.find(node2Id); i < 0 {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}
