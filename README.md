# gopeers
Gopeers is an implementation of Kademlia algorithm in golang. The KAD algorithm is a distributed hash
table (DHT) that is behind BitTorrent protocol. The current state of the project is that the core part
of Kademlia is implemented. Below is a list of next steps for the implementation.

# TODO
- periodic refresh
- caching
- key re-publishing
- replacement cache
- lock unresponsive peers (with backoff)
