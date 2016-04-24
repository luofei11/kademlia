package libkademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	table       RoutingTable
	data        map[ID][]byte
	updateChan chan Contact
}

func NewKademliaWithId(laddr string, nodeID ID) *Kademlia {
	k := new(Kademlia)
	k.NodeID = nodeID

	// TODO: Initialize other state here as you add functionality.
	k.table.Initialize()
	k.data = make(map[ID][]byte)
	k.updateChan = make(chan Contact)
	go k.HandleUpdate()
	// Set up RPC server
	// NOTE: KademliaRPC is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&KademliaRPC{k})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}
	s.HandleHTTP(rpc.DefaultRPCPath+hostname+port,
		rpc.DefaultDebugPath+hostname+port)
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// Run RPC server forever.
	go http.Serve(l, nil)

	// Add self contact
	hostname, port, _ = net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}
	return k
}

func NewKademlia(laddr string) *Kademlia {
	return NewKademliaWithId(laddr, NewRandomID())
}

type ContactNotFoundError struct {
	id  ID
	msg string
}
type ValueNotFoundError struct{
	key ID
}

func (e *ContactNotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}
func (e *ValueNotFoundError) Error() string {
	return fmt.Sprintf("Value not found for key: %x", e.key)
}

func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Self is target
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	}
	// Find contact with provided ID
	bucketIndex := k.FindBucket(nodeId)
	kbucket := k.table[bucketIndex]
	for _, contact := range kbucket {
		  if contact.NodeID.Equals(nodeId){
				  return &contact, nil
			}
	}
	return nil, &ContactNotFoundError{nodeId, "Not found"}
}
func (k *Kademlia) FindBucket(nodeId ID) int{
	//find the bucket the node falls into, return the index
  if k.NodeID.Equals(nodeId){
		  return -1
	}
	return (IDBits - 1) - k.NodeID.Xor(nodeId).PrefixLen()
}

type CommandFailed struct {
	msg string
}

func (e *CommandFailed) Error() string {
	return fmt.Sprintf("%s", e.msg)
}

func (k *Kademlia) DoPing(host net.IP, port uint16) (*Contact, error) {
	// TODO: Implement
  addr := fmt.Sprintf("%v:%v", host, port)
	port_str := fmt.Sprintf("%v", port)
	path := rpc.DefaultRPCPath + port_str
  client, err := rpc.DialHTTPPath("tcp", addr, path)
	if err != nil{
		  fmt.Println("Im here")
		  return nil, &CommandFailed{
				"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
	}
	fmt.Println("passed 1")
	defer client.Close()
  ping := PingMessage{k.SelfContact, NewRandomID()}
	var pong PongMessage
	err = client.Call("KademliaRPC.Ping", ping, &pong)
	if err != nil{
		  return nil, err
	}
	k.Update(pong.Sender)
	return &pong.Sender, nil
	//return nil, &CommandFailed{
		//"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) error {
	// TODO: Implement
	return &CommandFailed{"Not implemented"}
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) ([]Contact, error) {
	// TODO: Implement
	return nil, &CommandFailed{"Not implemented"}
}

func (k *Kademlia) DoFindValue(contact *Contact,
	searchKey ID) (value []byte, contacts []Contact, err error) {
	// TODO: Implement
	return nil, nil, &CommandFailed{"Not implemented"}
}

func (k *Kademlia) LocalFindValue(searchKey ID) ([]byte, error) {
	// TODO: Implement
	if val, ok := k.data[searchKey]; ok{
		return val, nil
	} else{
		return []byte(""), &ValueNotFoundError{searchKey}
	}
}

func (k *Kademlia) Update(c Contact) {
  //Update KBucket in Routing Table by Contact c
  k.updateChan <- c
}

func (k *Kademlia) HandleUpdate() {
	for {
		c := <- k.updateChan
		bucketIndex := k.FindBucket(c.NodeID)
		kb := k.table[bucketIndex]
		contains, i := kb.FindContactInKBucket(c)
		if contains {
			kb.MoveToTail(i)
		} else {
				if len(kb) < cap(kb) {
					kb.AddToTail(c)
				} else {
					head := kb[0]
					_, err := k.DoPing(head.Host, head.Port)
					if err != nil {
						kb.Remove(0)
						kb.AddToTail(c)
					}
				}
		}
	}
}

// For project 2!
func (k *Kademlia) DoIterativeFindNode(id ID) ([]Contact, error) {
	return nil, &CommandFailed{"Not implemented"}
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]Contact, error) {
	return nil, &CommandFailed{"Not implemented"}
}
func (k *Kademlia) DoIterativeFindValue(key ID) (value []byte, err error) {
	return nil, &CommandFailed{"Not implemented"}
}

// For project 3!
func (k *Kademlia) Vanish(data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	return
}

func (k *Kademlia) Unvanish(searchKey ID) (data []byte) {
	return nil
}
