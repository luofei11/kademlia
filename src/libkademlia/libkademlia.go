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
// Key value pair of data
type KVPair struct {
	key ID
	value []byte
}

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	table       RoutingTable
	data        map[ID][]byte
	channel     KademliaChannel
}
// KademliaChannel type used for communications
type KademliaChannel struct{
	updateChan chan Contact
	updateFinishedChan chan bool
	storeDataChan chan *KVPair
	valueLookUpChan chan ID
	valLookUpResChan chan []byte
}
func (kc *KademliaChannel) Initialize(){
	kc.updateChan = make(chan Contact)
	kc.updateFinishedChan = make(chan bool)
	kc.storeDataChan = make(chan *KVPair)
	kc.valueLookUpChan = make(chan ID)
	kc.valLookUpResChan = make(chan []byte)
}

func NewKademliaWithId(laddr string, nodeID ID) *Kademlia {
	k := new(Kademlia)
	k.NodeID = nodeID

	// TODO: Initialize other state here as you add functionality.
	k.table.Initialize()
	k.data = make(map[ID][]byte)
	k.channel.Initialize()
	go k.HandleUpdate()
	go k.HandleDataStore()
	go k.HandleValueLookUp()
	// Set up RPC server
	// NOTE: KademliaRPC is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&KademliaRPC{k})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}
	s.HandleHTTP(rpc.DefaultRPCPath+port,
		rpc.DefaultDebugPath+port)
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
//////////////////////////////////
//Error types
//////////////////////////////////
type ContactNotFoundError struct {
	id  ID
	msg string
}
type ValueNotFoundError struct{
	key ID
}
type CommandFailed struct {
	msg string
}
func (e *ContactNotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}
func (e *ValueNotFoundError) Error() string {
	return fmt.Sprintf("Value not found for key: %x", e.key)
}
func (e *CommandFailed) Error() string {
	return fmt.Sprintf("%s", e.msg)
}


func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Self is target
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	}
	// Find contact with provided ID
	bucketIndex := k.FindBucket(nodeId)
	if bucketIndex == -1 {
		return nil, &ContactNotFoundError{nodeId, "Not found"}
	}
	kbucket := k.table[bucketIndex]
	for _, contact := range kbucket {
		  if contact.NodeID.Equals(nodeId){
				  return &contact, nil
			}
	}
	return nil, &ContactNotFoundError{nodeId, "Not found"}
}

//////////////////////////////////////////////////////
//Doing corresponding RPC calls
//////////////////////////////////////////////////////
func (k *Kademlia) DoPing(host net.IP, port uint16) (*Contact, error) {
	// TODO: Implement
	addr := fmt.Sprintf("%v:%v", host, port)
	port_str := fmt.Sprintf("%v", port)
	path := rpc.DefaultRPCPath + port_str
	client, err := rpc.DialHTTPPath("tcp", addr, path)
	if err != nil{
		  return nil, &CommandFailed{
				"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
	}
	defer client.Close()
	ping := PingMessage{k.SelfContact, NewRandomID()}
	var pong PongMessage
	err = client.Call("KademliaRPC.Ping", ping, &pong)
	if err != nil{
		  return nil, err
	}

	k.Update(pong.Sender)
	return &pong.Sender, nil

}
func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) error {
	// TODO: Implement
	addr := fmt.Sprintf("%v:%v", (*contact).Host, (*contact).Port)
	port_str := strconv.Itoa(int((*contact).Port))
	path := rpc.DefaultRPCPath + port_str
	client, err := rpc.DialHTTPPath(
		"tcp",
		addr,
		path,
	)
	if err != nil {
		return err
	}
	defer client.Close()

	req := StoreRequest{k.SelfContact, NewRandomID(), key, value}
	var res StoreResult

	err = client.Call("KademliaRPC.Store", req, &res)
	//fmt.Println("dostore reaches here step6 !")
	if err != nil {
		client.Close()
		return err
	}
	return nil
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) ([]Contact, error) {
	// TODO: Implement
	addr := fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port)
	port_str := strconv.Itoa(int((*contact).Port))
	client, err := rpc.DialHTTPPath(
		"tcp",
		addr,
		rpc.DefaultRPCPath+port_str,
	)
	if err != nil {
		return  nil, err
	}
	defer client.Close()
	req := FindNodeRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindNodeResult
	err = client.Call("KademliaRPC.FindNode", req, &res)
	if err != nil {
		client.Close()
		return nil ,err
	}
	for _, each := range res.Nodes {
		k.Update(each)
	}
	return res.Nodes, nil
}

func (k *Kademlia) DoFindValue(contact *Contact,
	searchKey ID) (value []byte, contacts []Contact, err error) {
	// TODO: Implement
	addr := fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port)
	port_str := strconv.Itoa(int((*contact).Port))
	path := rpc.DefaultRPCPath + port_str
	client, err := rpc.DialHTTPPath(
		"tcp",
		addr,
		path,
	)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()
	req := FindValueRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindValueResult
	err = client.Call("KademliaRPC.FindValue", req, &res)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	if res.Value != nil {
		return res.Value, res.Nodes, nil
	} else if res.Nodes != nil {
		for _, node := range res.Nodes {
			k.Update(node)
		}
		return res.Value, res.Nodes, nil
	} else {
		return nil, nil, &CommandFailed{"Value Not Found"}
	}
	return nil, nil, &CommandFailed{"Value Not Found"}
}


///////////////////////////////////////////
//Interfaces of kademlia
///////////////////////////////////////////
func (k * Kademlia) StoreData(pair *KVPair){
	k.channel.storeDataChan <- pair
}
func (k *Kademlia) Update(c Contact) {
  //Update KBucket in Routing Table by Contact c
	k.channel.updateChan <- c
	_ = <- k.channel.updateFinishedChan
}
func (k *Kademlia) LookUpValue(key ID) ([]byte, error){
	//TODO: add lookup request to channel
	k.channel.valueLookUpChan <- key
	valLookUpResult := <- k.channel.valLookUpResChan
	if valLookUpResult != nil{
		  return valLookUpResult, nil
	}else{
		  return nil, &ValueNotFoundError{key}
	}
}


///////////////////////////////////////////
//Channel handlers of kademlia
///////////////////////////////////////////
func (k *Kademlia) HandleDataStore(){
	for {
		kvpair := <- k.channel.storeDataChan
		k.data[kvpair.key] = kvpair.value
	}
}
func (k *Kademlia) HandleValueLookUp(){
	for {
		key := <- k.channel.valueLookUpChan
		val, err := k.LocalFindValue(key)
		if err != nil{
			k.channel.valLookUpResChan <- nil
			}else{
				k.channel.valLookUpResChan <- val
			}
		}
}
func (k *Kademlia) HandleUpdate() {
	for {
		c := <- k.channel.updateChan
		//fmt.Println("New Contact to Update:",c)
		//fmt.Println("Original Kademlia:", k)
		bucketIndex := k.FindBucket(c.NodeID)
		if bucketIndex == -1 {
			k.channel.updateFinishedChan <- true
			continue
		}
		kb := &k.table[bucketIndex]
		contains, i := kb.FindContactInKBucket(c)
		if contains {
			kb.MoveToTail(i)
		} else {
			if len(*kb) < cap(*kb) {
				kb.AddToTail(c)
				} else {
					//fmt.Println("filled")
					head := (*kb)[0]
					_, err := k.DoPing(head.Host, head.Port)
					if err != nil {
						kb.Remove(0)
						kb.AddToTail(c)
						} else {
							kb.MoveToTail(0)
						}
					}
				}
				//fmt.Println("Updated kbucket:", kb)
				//fmt.Println("Updated Kademlia:", k)
				k.channel.updateFinishedChan <- true
			}
}

///////////////////////////////////////////////
//Helper functions
///////////////////////////////////////////////
func (k *Kademlia) LocalFindValue(searchKey ID) ([]byte, error) {
	// TODO: Implement
	if val, ok := k.data[searchKey]; ok{
		return val, nil
	} else{
		return nil, &ValueNotFoundError{searchKey}
	}
}
func (k *Kademlia) FindClosest(key ID) []Contact {
	prefixLen := k.NodeID.Xor(key).PrefixLen()
	var index int
	if prefixLen == 160 {
		index = 0
		} else {
		index = 159 - prefixLen
	}
	contacts := make([]Contact, 0, 20)
	for _, val := range k.table[index] {
		if len(contacts) < 20 {
			contacts = append(contacts, val)
		} else {
			return contacts
		}
	}

	if len(contacts) >= 20 {
		return contacts
	}

	//If the target kbucket has less than k contact, search higher bucket first, then lower bucket
	higher := index
	lower := index
	for  {
		if len(contacts) >= 20 {
			return contacts
		}
		if higher < 159 {
			higher++
			for _, val := range k.table[higher] {
				if len(contacts) < 20 {
					contacts = append(contacts, val)
				} else {
					return contacts
				}
			}
		}
		if lower > 0 {
			lower--
			for _, val := range k.table[lower] {
				if len(contacts) < 20 {
					contacts = append(contacts, val)
				} else {
					return contacts
				}
			}
		}
		if higher == 159 && lower == 0 {
			return contacts
		}
	}
	return contacts
}
func (k *Kademlia) FindBucket(nodeId ID) int{
	//find the bucket the node falls into, return the index
	if k.NodeID.Equals(nodeId){
		return -1
	}
	return (IDBits - 1) - k.NodeID.Xor(nodeId).PrefixLen()
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
