# gopeers
Go implementation of Kademlia algorithm.


## Lookup
- find the closest nodes to given id in local routing table
- send parallel rpc lookup calls to 'a' (eg a=3) closest nodes
- send parallel rpc lookup calls to 'a' closest nodes returned from previous
  rpc calls that haven't been queried yet
- if rpc calls are not returning nodes any closer,
  then send rpc calls to nodes that haven't been queried so far
  and are within k closest nodes
- terminate lookup when k closest nodes where queried successfully