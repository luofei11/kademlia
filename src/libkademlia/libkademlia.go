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
	"sort"
	"time"
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
type ShortListElement struct {
	contact Contact
	distance int
	status int//0 default, 1 inactive, 2 active
	hasValue bool
}
type ShortListElements []ShortListElement
func (slice ShortListElements) Len() int {
	return len(slice)
}
func (slice ShortListElements) Less(i, j int) bool {
	return slice[i].distance < slice[j].distance
}
func (slice ShortListElements) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func NotEnoughActive(ContactedList []ShortListElement) bool {
	count_active := 0
	for _, val := range ContactedList {
		if val.status == 2 {
			count_active++
		}
	}
	if count_active < 20 {
		return true
	} else {
		return false
	}
}

type IterFindNodeResult struct {
	Receiver Contact
	Nodes []Contact
	Err   error
}

type IterFindValueResult struct {
	receiver Contact
	val []byte
	contacts []Contact
	err error
}

func (k *Kademlia) iteFindNodeHelper(server ShortListElement, id ID, iterFindNodeChan chan IterFindNodeResult) {
	var res IterFindNodeResult
	res.Receiver = server.contact
	res.Nodes, res.Err = k.DoFindNode(&server.contact, id)
	iterFindNodeChan <- res
}
func (k *Kademlia) iterFindValueHelper(server ShortListElement, id ID, iterFindValueChan chan IterFindValueResult) {
	var res IterFindValueResult
	res.receiver = server.contact
	res.val, res.contacts, res.err = k.DoFindValue(&server.contact, id)
	iterFindValueChan <- res
}

func notInList(List []ShortListElement, one_shortlist_element ShortListElement) bool {
	for _, val := range List {
		if val.contact.NodeID.Equals(one_shortlist_element.contact.NodeID) {
			return false
		}
	}
	return true
}

func (k *Kademlia) DoIterativeFindNode(id ID) ([]Contact, error) {
	ShortList := make([]ShortListElement, 0, 60)
	ProbingList := make([]ShortListElement, 0, 3)
	ContactedList := make([]ShortListElement, 0, 30)
	iterFindNodeChan := make(chan IterFindNodeResult)

	initial_shortlist := k.FindClosest(id)
	for _, val := range initial_shortlist {
		one_shortlist_element := ShortListElement{val, 159 - id.Xor(val.NodeID).PrefixLen(), 0, false}
		ShortList = append(ShortList, one_shortlist_element)
	}
	sort.Sort(ShortListElements(ShortList))

	ClosestDistancePre := 160
	ClosestDistanceNow := 159
	if len(ShortList) > 0 {
		ClosestDistanceNow = ShortList[0].distance
	}

	for (ClosestDistanceNow < ClosestDistancePre && NotEnoughActive(ContactedList)){
		ClosestDistancePre = ClosestDistanceNow
		ProbingList = nil
		//dump(ProbingList)
		i := 0
		for i < 3 && i < len(ShortList) {
			ProbingList = append(ProbingList, ShortList[i])
			i++
		}
		ShortList = append(ShortList[:0], ShortList[i:]...)
		timeout := false
		timeOutChan := make(chan bool)
		go func() {
			time.Sleep(300 * time.Millisecond)
			timeOutChan <- true
		}()

		for _, val := range ProbingList {
			go k.iteFindNodeHelper(val, id, iterFindNodeChan)
		}
		allreceive := false

		for !timeout && !allreceive{
			select {
				case res := <- iterFindNodeChan:
					for index, val := range ProbingList {
						if val.contact.NodeID.Equals(res.Receiver.NodeID){
							ProbingList = append(ProbingList[:index], ProbingList[index + 1:]...)
							one_shortlist_element := ShortListElement{res.Receiver, 159 - id.Xor(res.Receiver.NodeID).PrefixLen(), 0, false}
							if res.Err != nil {
								one_shortlist_element.status = 1
							} else {
								one_shortlist_element.status = 2
							}
							ContactedList = append(ContactedList, one_shortlist_element)
						}
					}
					// inContactedList := false
					// for _, val := range ContactedList {
					// 	if val.contact.NodeID.Equals(res.Receiver.NodeID) {
					// 		if res.Err != nil {
					// 			val.status = 1
					// 		} else {
					// 			val.status = 2
					// 		}
					// 		inContactedList = true
					// 	}
					// }
					// if !inContactedList {
					// 	one_shortlist_element := ShortListElement{res.Receiver, 159 - id.Xor(res.Receiver.NodeID).PrefixLen(), 0}
					// 	if res.Err != nil {
					// 		one_shortlist_element.status = 1
					// 	} else {
					// 		one_shortlist_element.status = 2
					// 	}
					// 	ContactedList = append(ContactedList, one_shortlist_element)
					// }
					if res.Err == nil {
						for _, val := range res.Nodes {
							one_shortlist_element := ShortListElement{val, 159 - id.Xor(val.NodeID).PrefixLen(), 0, false}
							if notInList(ShortList, one_shortlist_element) && notInList(ProbingList, one_shortlist_element) && notInList(ContactedList, one_shortlist_element){
								ShortList = append(ShortList, one_shortlist_element)
							}
						}
					}
					if len(ProbingList) == 0 {
						allreceive = true
					}
				case timeout =<- timeOutChan:
					for _, probingval := range ProbingList {
						// inContactedList := false
						// for _, contactedval := range ContactedList {
						// 	if probingval.contact.NodeID.Equals(contactedval.contact.NodeID) {
						// 		contactedval.status = 1
						// 		inContactedList = true
						// 	}
						// }
						// if !inContactedList {
							probingval.status = 1
							ContactedList = append(ContactedList, probingval)
						// }
					}
			}
		}
		sort.Sort(ShortListElements(ShortList))
		if len(ShortList) > 0 {
			ClosestDistanceNow = ShortList[0].distance
		}
	}

	for (NotEnoughActive(ContactedList) && len(ShortList) > 0){
		ProbingList = nil
		//dump(ProbingList)
		i := 0
		for i < 3 && i < len(ShortList) {
			ProbingList = append(ProbingList, ShortList[i])
			i++
		}
		ShortList = append(ShortList[:0], ShortList[i:]...)
		timeout := false
		timeOutChan := make(chan bool)
		go func() {
			time.Sleep(300 * time.Millisecond)
			timeOutChan <- true
		}()

		for _, val := range ProbingList {
			go k.iteFindNodeHelper(val, id, iterFindNodeChan)
		}
		allreceive := false

		for !timeout && !allreceive{
			select {
				case res := <- iterFindNodeChan:
					for index, val := range ProbingList {
						if val.contact.NodeID.Equals(res.Receiver.NodeID){
							ProbingList = append(ProbingList[:index], ProbingList[index + 1:]...)
							one_shortlist_element := ShortListElement{res.Receiver, 159 - id.Xor(res.Receiver.NodeID).PrefixLen(), 0, false}
							if res.Err != nil {
								one_shortlist_element.status = 1
							} else {
								one_shortlist_element.status = 2
							}
							ContactedList = append(ContactedList, one_shortlist_element)
						}
					}
					// inContactedList := false
					// for _, val := range ContactedList {
					// 	if val.contact.NodeID.Equals(res.Receiver.NodeID) {
					// 		if res.Err != nil {
					// 			val.status = 1
					// 		} else {
					// 			val.status = 2
					// 		}
					// 		inContactedList = true
					// 	}
					// }
					// if !inContactedList {
					// 	one_shortlist_element := ShortListElement{res.Receiver, 159 - id.Xor(res.Receiver.NodeID).PrefixLen(), 0}
					// 	if res.Err != nil {
					// 		one_shortlist_element.status = 1
					// 	} else {
					// 		one_shortlist_element.status = 2
					// 	}
					// 	ContactedList = append(ContactedList, one_shortlist_element)
					// }
					if res.Err == nil {
						for _, val := range res.Nodes {
							one_shortlist_element := ShortListElement{val, 159 - id.Xor(val.NodeID).PrefixLen(), 0, false}
							if notInList(ShortList, one_shortlist_element) && notInList(ProbingList, one_shortlist_element) && notInList(ContactedList, one_shortlist_element){
								ShortList = append(ShortList, one_shortlist_element)
							}
						}
					}
					if len(ProbingList) == 0 {
						allreceive = true
					}
				case timeout =<- timeOutChan:
					for _, probingval := range ProbingList {
						// inContactedList := false
						// for _, contactedval := range ContactedList {
						// 	if probingval.contact.NodeID.Equals(contactedval.contact.NodeID) {
						// 		contactedval.status = 1
						// 		inContactedList = true
						// 	}
						// }
						// if !inContactedList {
							probingval.status = 1
							ContactedList = append(ContactedList, probingval)
						// }
					}
			}
		}
		sort.Sort(ShortListElements(ShortList))
	}

	ResultList := make([]Contact, 0, 30)
	sort.Sort(ShortListElements(ContactedList))
	for _, val := range ContactedList {
		if val.status == 2 {
			ResultList = append(ResultList, val.contact)
		}
	}
	if len(ResultList) > 20 {
		ResultList = ResultList[:20]
	}
	return ResultList, nil
	//return nil, &CommandFailed{"Not implemented"}
}

func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]Contact, error) {
	contacts, _:= k.DoIterativeFindNode(key)
	ResultList := make([]Contact, 0, 30)
	for _, con := range contacts {
		errormsg := k.DoStore(&con, key, value)
		if errormsg == nil {
			ResultList = append(ResultList, con)
		}
	}
	return ResultList, nil
	//return nil, &CommandFailed{"Not implemented"}
}
func (k *Kademlia) DoIterativeFindValue(key ID) (value []byte, err error) {
	ShortList := make([]ShortListElement, 0, 60)
	ProbingList := make([]ShortListElement, 0, 3)
	ContactedList := make([]ShortListElement, 0, 30)
	iterFindValueChan := make(chan IterFindValueResult, 3)

	initial_shortlist := k.FindClosest(key)
	for _, val := range initial_shortlist {
		one_shortlist_element := ShortListElement{val, 159 - key.Xor(val.NodeID).PrefixLen(), 0, false}
		ShortList = append(ShortList, one_shortlist_element)
	}
	sort.Sort(ShortListElements(ShortList))

	ClosestDistancePre := 160
	ClosestDistanceNow := 159
	if len(ShortList) > 0 {
		ClosestDistanceNow = ShortList[0].distance
	}
	valueFound := false
	var finalValue []byte = nil

	for (ClosestDistanceNow < ClosestDistancePre) && NotEnoughActive(ContactedList) && (!valueFound) {
		ClosestDistancePre = ClosestDistanceNow
		ProbingList = nil
		//dump(ProbingList)
		i := 0
		for i < alpha && i < len(ShortList) {
			ProbingList = append(ProbingList, ShortList[i])
			i++
		}
		ShortList = append(ShortList[:0], ShortList[i:]...)
		timeout := false
		timeOutChan := make(chan bool)
		//valueFoundChan := make(chan bool)
		//finalValueChan := make(chan []byte)
		go func() {
			time.Sleep(300 * time.Millisecond)
			timeOutChan <- true
		}()

		for _, val := range ProbingList {
			go k.iterFindValueHelper(val, key, iterFindValueChan)
		}
		allreceive := false
		for !timeout && !allreceive {
			select {
			case res := <- iterFindValueChan:
				  //fmt.Println("I received something:", res.val)
					for index, val := range ProbingList {
						if val.contact.NodeID.Equals(res.receiver.NodeID) {
							ProbingList = append(ProbingList[:index], ProbingList[index + 1:]...)
							one_shortlist_element := ShortListElement{res.receiver, 159 - key.Xor(res.receiver.NodeID).PrefixLen(), 0, false}
							if res.err != nil {
								one_shortlist_element.status = 1
							} else {
								one_shortlist_element.status = 2
							}
							if res.val != nil {
								one_shortlist_element.hasValue = true
								valueFound = true
								finalValue = res.val
								//fmt.Println("final Value is :", finalValue)
							} else {
								one_shortlist_element.hasValue = false
								fmt.Println("didn't find value")
								fmt.Println("finalValue is:", finalValue)
							}
							ContactedList = append(ContactedList, one_shortlist_element)
						}
					}

					if res.err == nil {
						for _, val := range res.contacts {
							one_shortlist_element := ShortListElement{val, 159 - key.Xor(val.NodeID).PrefixLen(), 0, false}
							if notInList(ShortList, one_shortlist_element) && notInList(ProbingList, one_shortlist_element) && notInList(ContactedList, one_shortlist_element) {
								ShortList = append(ShortList, one_shortlist_element)
							}
						}
					}
					if len(ProbingList) == 0 {
						allreceive = true
						//fmt.Println("All received")
					}

				case timeout= <- timeOutChan:
					//fmt.Println("timeout!!!!!!")
					for _, probingval := range ProbingList {
							probingval.status = 1
							probingval.hasValue = false
							ContactedList = append(ContactedList, probingval)
					}
			}
		}
		sort.Sort(ShortListElements(ShortList))
		if len(ShortList) > 0 {
			ClosestDistanceNow = ShortList[0].distance
		}
	}
	sort.Sort(ShortListElements(ContactedList))
  close(iterFindValueChan)

  if valueFound {
		  //fmt.Println("I found value1: !", finalValue)
			for _, con := range ContactedList{
				if (con.status == 2 && !con.hasValue) {
					k.DoStore(&con.contact, key, finalValue)
				}
			}
			return finalValue, nil
	} else {
		  //fmt.Println("Did I enter this condition1?")
      return nil, &ValueNotFoundError{ContactedList[0].contact.NodeID}
	}
	return nil, &ValueNotFoundError{ContactedList[0].contact.NodeID}
	//return nil, &CommandFailed{"Not implemented"}
}


// For project 3!
func (k *Kademlia) Vanish(data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	return
}

func (k *Kademlia) Unvanish(searchKey ID) (data []byte) {
	return nil
}
