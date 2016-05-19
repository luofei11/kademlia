package libkademlia

import (
	//"bytes"
	//"net"
	"strconv"
	"testing"
	"sort"
	"time"
	"fmt"
	"reflect"
	//"container/heap"
)

func GenerateTreeKademlia(num_treenode int, start_port int) []*Kademlia {
	ResultList := make([]*Kademlia, 0, num_treenode)
	root_address := "localhost:" + strconv.Itoa(start_port)
	root_kademlia := NewKademlia(root_address)
	ResultList = append(ResultList, root_kademlia)
	for i := 1; i < num_treenode; i++ {
		leaf_address := "localhost:" + strconv.Itoa(start_port + i)
		leaf_kademlia := NewKademlia(leaf_address)
		ResultList = append(ResultList, leaf_kademlia)
		father_address := "localhost:" + strconv.Itoa(start_port + i / 3)
		host_number, port_number, _ := StringToIpPort(father_address)
		ResultList[i].DoPing(host_number, port_number)
	}
	return ResultList
}

// func GenerateTreeIDList(num int) (ret []ID) {
// 	ret = make([]ID, num)
// 	ret[0] = NewRandomID()
// 	for i := 1; i < num; i++ {
// 		if i > 150 {
// 			ret[i] = NewRandomID()
// 		} else {
// 			curID := ret[i/divNum]
// 			curID[i/8] = curID[i/8] ^ (1 << uint8(7-(i%8)))
// 			ret[i] = curID
// 		}
// 	}
// 	return ret
// }

// func GenerateTestList(num int, idList []ID) (kRet KademliaList, cRet []Contact) {
// 	kRet = []*Kademlia{}
// 	cRet = []Contact{}
// 	for i := 0; i < num; i++ {
// 		laddr := testAddr + ":" + strconv.Itoa(int(testPort))
// 		testPort++
// 		var k *Kademlia
// 		if idList != nil && i < len(idList) {
// 			k = NewKademliaWithId(laddr, idList[i])
// 		} else {
// 			k = NewKademlia(laddr)
// 		}
// 		cRet = append(cRet, k.SelfContact)
// 		kRet = append(kRet, k)
// 	}
// 	return
// }
// var testPort uint16 = 3000
//
// const testAddr = "localhost"
// const divNum = 3
//
// type KademliaList []*Kademlia
//
// func GenerateRandomIDList(num int) (ret []ID) {
// 	ret = make([]ID, num)
// 	for i := 0; i < num; i++ {
// 		ret[i] = NewRandomID()
// 	}
// 	return
// }
//
// func GenerateTreeIDList(num int) (ret []ID) {
// 	ret = make([]ID, num)
// 	ret[0] = NewRandomID()
// 	for i := 1; i < num; i++ {
// 		if i > 150 {
// 			ret[i] = NewRandomID()
// 		} else {
// 			curID := ret[i/divNum]
// 			curID[i/8] = curID[i/8] ^ (1 << uint8(7-(i%8)))
// 			ret[i] = curID
// 		}
// 	}
// 	return ret
// }
//
// func GenerateTestList(num int, idList []ID) (kRet KademliaList, cRet []Contact) {
// 	kRet = []*Kademlia{}
// 	cRet = []Contact{}
// 	for i := 0; i < num; i++ {
// 		laddr := testAddr + ":" + strconv.Itoa(int(testPort))
// 		testPort++
// 		var k *Kademlia
// 		if idList != nil && i < len(idList) {
// 			k = NewKademliaWithId(laddr, idList[i])
// 		} //else {
// 			//k = NewKademliaWithId(laddr, nil)
// 		//}
// 		cRet = append(cRet, k.SelfContact)
// 		kRet = append(kRet, k)
// 	}
// 	return
// }
//
// func (ks KademliaList) ConnectTo(k1, k2 int) {
// 	ks[k1].DoPing(ks[k2].SelfContact.Host, ks[k2].SelfContact.Port)
// }
//
// func SortContact(input []Contact, key ID) (ret []Contact) {
// 	cHeap := &ContactHeap{input, key}
// 	heap.Init(cHeap)
// 	ret = []Contact{}
// 	for cHeap.Len() > 0 {
// 		ret = append(ret, heap.Pop(cHeap).(Contact))
// 	}
// 	return
// }

func TestIterativeFindNodeSimple(t *testing.T) {
	//Structure: 1 - 2 - 3
	//Simple Test: Do Interative Find Node 3 in 1
	//Should return all 1, 2, 3
	instance1 := NewKademlia("localhost:7950")
	instance2 := NewKademlia("localhost:7951")
	instance3 := NewKademlia("localhost:7952")

	host2, port2, _ := StringToIpPort("localhost:7951")
	instance1.DoPing(host2, port2)
	host3, port3, _ := StringToIpPort("localhost:7952")
	instance2.DoPing(host3, port3)

	contacts, err := instance1.DoIterativeFindNode(instance3.NodeID)
	if err != nil {
		t.Error("Error doing IterativeFindNode")
	}
	if contacts == nil || len(contacts) == 0 {
		t.Error("No contacts were found")
	}
	find1 := false
	find2 := false
	find3 := false
	for _, val := range contacts {
		if val.NodeID.Equals(instance1.NodeID) {
			find1 = true
		} else if val.NodeID.Equals(instance2.NodeID) {
			find2 = true
		} else if val.NodeID.Equals(instance3.NodeID) {
			find3 = true
		}
	}
	if !find1 || !find2 || !find3 {
		t.Error("Didn't Find All in DoIterativeFindNode!")
	}
}

func TestIterativeFindNode(t *testing.T) {
	//Tree Structure:
	/*
	                          0
	             /                         \
	            1                          2
	     /      |        \         /       |        \
	   3        4        5        6        7        8
	/  |  \  /  |  \  /  |  \  /  |  \  /  |  \  /  |  \
	9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26
	*/
	//Do Iterative Find Node in Node 0. It should be able to find anyone in this tree.
	num_treenode := 27
	tree_kademlia := GenerateTreeKademlia(num_treenode, 8000)
	for i := 1; i < num_treenode; i++ {
		search_ID := tree_kademlia[i].NodeID
		result_list, err := tree_kademlia[0].DoIterativeFindNode(search_ID)
		if err != nil {
			t.Error("DoIterativeFindNode Return Error: ", i)
		}
		result_list_for_sort := make([]ShortListElement, 0, 20)
		for _, val := range result_list {
			one_shortlist_element := ShortListElement{val, 159 - search_ID.Xor(val.NodeID).PrefixLen(), 0, false}
			result_list_for_sort = append(result_list_for_sort, one_shortlist_element)
		}
		sort.Sort(ShortListElements(result_list_for_sort))
		if !result_list_for_sort[0].contact.NodeID.Equals(search_ID) {
			t.Error("DoIterativeFindNode Doesn't Find Search_ID: ", i)
		}
	}
	return
}
func TestIterativeFindValue(t *testing.T) {
	num_treenode := 27
	tree_kademlia := GenerateTreeKademlia(num_treenode, 8050)
	time.Sleep(100 * time.Millisecond)
	searchKey := tree_kademlia[num_treenode - 1].NodeID
	searchValue := []byte("hello world!")
	tree_kademlia[(num_treenode - 1) / 3].DoStore(&tree_kademlia[num_treenode - 1].SelfContact, searchKey, searchValue)
	time.Sleep(100 * time.Millisecond)
	v, err := tree_kademlia[num_treenode - 1].LocalFindValue(searchKey)
	if err != nil {
		t.Error("DoStore error!")
		return
	}
	if string(v) != string(searchValue) {
		t.Error("Value doesn't match!")
		return
	}
	resultVal, err := tree_kademlia[0].DoIterativeFindValue(searchKey)
	if err != nil {
		t.Error("DoIterativeFindValue Return Error!")
		return
	}
	if string(resultVal) != string(searchValue) {
		t.Error("Value is not correct!")
		return
	}
	return
}
// func TestIterativeFindValueSimple(t *testing.T) {
// 	instance1 := NewKademlia("localhost:8939")
// 	instance2 := NewKademlia("localhost:8940")
// 	instance3 := NewKademlia("localhost:8941")
//
// 	host2, port2, _ := StringToIpPort("localhost:8940")
// 	instance1.DoPing(host2, port2)
// 	host3, port3, _ := StringToIpPort("localhost:8941")
// 	instance2.DoPing(host3, port3)
// 	searchKey := instance3.SelfContact.NodeID
// 	searchKey[IDBytes - 1] = 0
// 	searchValue := []byte("helloworld!")
// 	instance1.DoStore(&instance3.SelfContact, searchKey, searchValue)
// 	v, _ := instance2.DoIterativeFindValue(searchKey)
// 	if v == nil{
// 		t.Error("luofei test failed")
// 	}else{
// 		t.Log("luofei test succeed!")
// 	}
// 	return
// }
// func TestIterativeFindValue(t *testing.T) {
// 	idList = make([]ID, 20)
// 	idList[0] = NewRandomID()
// 	for i := 1; i < 20; i++ {
// 			curID := id[i / 3]
// 			curID[i/8] = curID[i/8] ^ (1 << uint8(7-(i%8)))
// 			ret[i] = curID
// 	}
// 	treeList := GenerateTreeIDList(kNum)
// 	kList, _ := GenerateTestList(kNum, treeList)
// 	for i := 1; i < kNum; i++ {
// 		kList.ConnectTo(i, i/divNum)
// 	}
// 	time.Sleep(100 * time.Millisecond)
// 	searchKey := kList[targetIdx].SelfContact.NodeID
// 	searchKey[IDBytes-1] = 0
// 	randValue := []byte(NewRandomID().AsString())
// 	kList[targetIdx/divNum].DoStore(&kList[targetIdx].SelfContact, searchKey, randValue)
// 	time.Sleep(3 * time.Millisecond)
// 	retVal, _ := kList[targetIdx].LocalFindValue(searchKey)
// 	if retVal == nil {
// 		t.Error("The target node should have the key/value pair")
// 		return
// 	}
// 	if string(retVal) != string(randValue) {
// 		t.Error("The stored value should equal to each other")
// 		return
// 	}
// 	kList[0].DoFindNode(&kList[targetIdx].SelfContact, searchKey)
// 	res, _ := kList[0].DoIterativeFindValue(searchKey)
// 	fmt.Println("I'm going to seach: ", searchKey, randValue)
// 	fmt.Println("Returned Value is: ", res)
// 	if res == nil {
// 		t.Error("The coressponding value should be found")
// 		return
// 	}
// 	if string(res) != string(randValue) {
// 		t.Error("Search result doesn't match: " + string(res) + "!=" + string(randValue))
// 	}
// 	t.Log("TestIterativeFindValue done successfully!\n")
// 	return
//
// }

func TestIterativeStore(t *testing.T) {
	instance1 := NewKademlia("localhost:3456")
	instance2 := NewKademlia("localhost:4567")
	instance3 := NewKademlia("localhost:5678")

	host2, port2, _ := StringToIpPort("localhost:4567")
	instance1.DoPing(host2, port2)
	host3, port3, _ := StringToIpPort("localhost:5678")
	instance2.DoPing(host3, port3)

	fmt.Println(instance1)
	fmt.Println(instance2)
	fmt.Println(instance3)

	key := instance3.NodeID
	value := []byte("hello")
	ContactList, err := instance1.DoIterativeStore(key, value)
	if err != nil {
		t.Error("Error doing IterativeStore")
	}

	contacts, _:= instance1.DoIterativeFindNode(key)
	testList := make([]Contact, 0, k)
	for _, con := range contacts {
		errormsg := instance1.DoStore(&con, key, value)
		if errormsg == nil {
			testList = append(testList, con)
		}
	}

	if reflect.DeepEqual(ContactList, testList) != true {
		t.Error("DoIterativeStore test fail.")
	}
}
// func TestIterativeFindValue(t *testing.T) {
// 	kNum := 50
// 	targetIdx := kNum - 23
// 	treeList := GenerateTreeIDList(kNum)
// 	kList, _ := GenerateTestList(kNum, treeList)
// 	for i := 1; i < kNum; i++ {
// 		kList.ConnectTo(i, i/divNum)
// 	}
// 	time.Sleep(100 * time.Millisecond)
// 	searchKey := kList[targetIdx].SelfContact.NodeID
// 	searchKey[IDBytes-1] = 0
// 	randValue := []byte(NewRandomID().AsString())
// 	kList[targetIdx/divNum].DoStore(&kList[targetIdx].SelfContact, searchKey, randValue)
// 	time.Sleep(3 * time.Millisecond)
// 	retVal, _ := kList[targetIdx].LocalFindValue(searchKey)
// 	if retVal == nil {
// 		t.Error("The target node should have the key/value pair")
// 		return
// 	}
// 	if string(retVal) != string(randValue) {
// 		t.Error("The stored value should equal to each other")
// 		return
// 	}
// 	res, _ := kList[0].DoIterativeFindValue(searchKey)
// 	fmt.Println("I'm going to seach: ", searchKey, randValue)
// 	fmt.Println("Returned Value is: ", res)
// 	if res == nil {
// 		t.Error("The coressponding value should be found")
// 		return
// 	}
// 	if string(res) != string(randValue) {
// 		t.Error("Search result doesn't match: " + string(res) + "!=" + string(randValue))
// 	}
// 	t.Log("TestIterativeFindValue done successfully!\n")
// 	return
//
// }
// func TestIterativeStore(t *testing.T) {
// 	instance1 := NewKademlia("localhost:3456")
// 	instance2 := NewKademlia("localhost:4567")
// 	instance3 := NewKademlia("localhost:5678")
//
// 	host2, port2, _ := StringToIpPort("localhost:4567")
// 	instance1.DoPing(host2, port2)
// 	host3, port3, _ := StringToIpPort("localhost:5678")
// 	instance2.DoPing(host3, port3)
//
// 	fmt.Println(instance1)
// 	fmt.Println(instance2)
// 	fmt.Println(instance3)
//
// 	key := instance3.NodeID
// 	value := []byte("hello")
// 	ContactList, err := instance1.DoIterativeStore(key, value)
// 	if err != nil {
// 		t.Error("Error doing IterativeStore")
// 	}
//
// 	contacts, _:= instance1.DoIterativeFindNode(key)
// 	testList := make([]Contact, 0, k)
// 	for _, con := range contacts {
// 		errormsg := instance1.DoStore(&con, key, value)
// 		if errormsg == nil {
// 			testList = append(testList, con)
// 		}
// 	}
//
// 	if reflect.DeepEqual(ContactList, testList) != true {
// 		t.Error("DoIterativeStore test fail.")
// 	}
// }
