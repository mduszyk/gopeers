# gopeers
This project implements the Kademlia distributed hash table (DHT) protocol in Go,
enabling efficient peer-to-peer lookup and storage in decentralized networks.
Kademlia is used in systems like BitTorrent's DHT and IPFS for its robustness,
scalability, and XOR-based routing.

The current state of the project is that the core part
of Kademlia is implemented. Below is a list of next steps for the implementation.

# TODO
- periodic refresh
- caching
- key re-publishing
- replacement cache
- lock unresponsive peers (with backoff)
