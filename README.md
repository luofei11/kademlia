# kademlia
![Kad](https://avatars1.githubusercontent.com/u/16706596?v=3&s=200)
===================================================================

Demo implementation of the
[Kademlia](http://www.scs.stanford.edu/~dm/home/papers/kpos.pdf) distributed
hash table written in [GoLang](https://golang.org/).


### BUILDING


> Go's build tools depend on the value of the GOPATH environment variable. $GOPATH
> should be the project root: the absolute path of the directory containing
> {bin,pkg,src}.

Once you've set that, you should be able to build the skeleton and create an
executable at bin/kademlia with:

    go install kademlia

Running it as

    kademlia localhost:7890 localhost:7890

will cause it to start up a server bound to localhost:7890 (the first argument)
and then connect as a client to itself (the second argument).


### COMMAND-LINE INTERFACE


* whoami
  - Print your node ID.

* local_find_value key
  - If your node has data for the given key, print it.
  - If your node does not have data for the given key, you should print "ERR".

* get_contact ID
  - If your buckets contain a node with the given ID,
        `printf("%v %v\n", theNode.addr, theNode.port)`
  - If your buckers do not contain any such node, print "ERR".

> The following four commands cause your code to invoke the appropriate RPC on another node, specified by the nodeID argument.

* ping nodeID
* ping host:port
  - Perform a ping.

* store nodeID key value
  - Perform a store and print a blank line.

* find_node nodeID key
  - Perform a find_node and print its results as for iterativeFindNode.

* find_value nodeID key
  - Perform a find_value. If it returns nodes, print them as for find_node. If it returns a value, print the value as in iterativeFindValue.

> The following commands are the iterative RPCs. These are for project 2.

* iterativeStore key value
  - Perform the iterativeStore operation and then print the ID of the node that
    received the final STORE operation.

* iterativeFindNode ID
  - Print a list of ≤ k closest nodes and print their IDs. You should collect
    the IDs in a slice and print that.

* iterativeFindValue key
  - `printf("%v %v\n", ID, value)`, where ID refers to the node that finally returned the value. If you do not find a value, print "ERR".
