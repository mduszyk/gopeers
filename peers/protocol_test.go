package peers

import (
	"testing"
	"time"
)

func TestUdpProtocol(t *testing.T) {

	nodePeer1 := Peer{LastSeen: time.Now()}
	buckets1 := NewBucketList(20, 5, nodePeer1)
	p2pNode1, err := NewP2pNode(buckets1)
	if err != nil {
		t.Errorf("failed creating p2p node: %v\n", err)
	}
	protoServer1, err := NewUdpProtocolServer("localhost:5001", p2pNode1)
	nodePeer1.Id = p2pNode1.id
	nodePeer1.Protocol = protoServer1.Connect(protoServer1.rpcNode.Addr)

	nodePeer2 := Peer{LastSeen: time.Now()}
	buckets2 := NewBucketList(20, 5, nodePeer2)
	p2pNode2, err := NewP2pNode(buckets2)
	if err != nil {
		t.Errorf("failed creating p2p node: %v\n", err)
	}
	protoServer2, err := NewUdpProtocolServer("localhost:5002", p2pNode2)
	nodePeer2.Id = p2pNode2.id
	nodePeer2.Protocol = protoServer1.Connect(protoServer2.rpcNode.Addr)

	proto1 := protoServer2.Connect(protoServer1.rpcNode.Addr)
	proto2 := protoServer1.Connect(protoServer2.rpcNode.Addr)

	sender := Peer{protoServer2.p2pNode.id, proto2, time.Now()}
	randomId, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id: %v\n", err)
	}
	echoId, err := proto1.Ping(sender, randomId)
	if err != nil {
		t.Errorf("failed pinging: %v\n", err)
	}
	if echoId.Cmp(randomId) != 0 {
		t.Errorf("ping returned invalid Id\n")
	}
	if i, _ := p2pNode1.buckets.find(p2pNode2.id); i < 0 {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}
