# gopeers
Gopeers is an implementation of Kademlia algorithm in go. The KAD algorithm is a distributed hash
table (DHT) that is behind BitTorrent protocol. The main goal of the project is an educational one
with an open possibility of pivoting into sth useful. In the current state the core part of Kademlia
is implemented. Below is a brief list of next steps for the implementation.

# TODO
- periodic refresh
- caching
- key re-publishing
- replacement cache
- lock unresponsive peers (with backoff)
