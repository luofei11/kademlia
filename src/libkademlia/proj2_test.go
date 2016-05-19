package libkademlia

import (
	//"bytes"
	//"net"
	"strconv"
	"testing"
	"sort"
	//"time"
	//"fmt"
	//"reflect"
	//"container/heap"
)

func GenerateTreeKademlia(num_treenode int) []*Kademlia {
	ResultList := make([]*Kademlia, 0, num_treenode)
	root_kademlia := NewKademlia("localhost:8000")
	ResultList = append(ResultList, root_kademlia)
	for i := 1; i < num_treenode; i++ {
		leaf_address := "localhost:" + strconv.Itoa(8000+i)
		leaf_kademlia := NewKademlia(leaf_address)
		ResultList = append(ResultList, leaf_kademlia)
		father_address := "localhost:" + strconv.Itoa(8000+i/3)
		host_number, port_number, _ := StringToIpPort(father_address)
		ResultList[i].DoPing(host_number, port_number)
	}
	return ResultList
}

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
	instance1 := NewKademlia("localhost:7950")
	instance2 := NewKademlia("localhost:7951")
	instance3 := NewKademlia("localhost:7952")

	host2, port2, _ := StringToIpPort("localhost:7951")
	instance1.DoPing(host2, port2)
	host3, port3, _ := StringToIpPort("localhost:7952")
	instance2.DoPing(host3, port3)

	// fmt.Println(instance1)
	// fmt.Println(instance2)
	// fmt.Println(instance3)

	contacts, err := instance1.DoIterativeFindNode(instance3.NodeID)
	if err != nil {
		t.Error("Error doing IterativeFindNode")
	}
	if contacts == nil || len(contacts) == 0 {
		t.Error("No contacts were found")
	}
	// fmt.Println(contacts)
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
	num_treenode := 60
	tree_kademlia := GenerateTreeKademlia(num_treenode)
	search_ID := tree_kademlia[num_treenode - 1].NodeID
	result_list, err := tree_kademlia[0].DoIterativeFindNode(search_ID)
	if err != nil {
		t.Error("DoIterativeFindNode Return Error!")
	}
	result_list_for_sort := make([]ShortListElement, 0, 20)
	for _, val := range result_list {
		one_shortlist_element := ShortListElement{val, 159 - search_ID.Xor(val.NodeID).PrefixLen(), 0, false}
		result_list_for_sort = append(result_list_for_sort, one_shortlist_element)
	}
	sort.Sort(ShortListElements(result_list_for_sort))
	if !result_list_for_sort[0].contact.NodeID.Equals(search_ID) {
		t.Error("DoIterativeFindNode Doesn't Find Search_ID!")
	}
	return

	// kNum := 60
	// targetIdx := kNum - 1
	// treeList := GenerateTreeIDList(kNum)
	// fmt.Println("treeList:", treeList)
	// kList, _ := GenerateTestList(kNum, treeList)
	// fmt.Println("kList:", kList)
	// for i := 1; i < kNum; i++ {
	// 	kList.ConnectTo(i, i/divNum)
	// }
	// fmt.Println("KList After Connect:", kList)
	// time.Sleep(100 * time.Millisecond)
	// searchKey := kList[targetIdx].SelfContact.NodeID
	// searchKey[IDBytes-1] = 0
	// fmt.Println("searchKey:", searchKey)
	// // res_Nodes, res_Err := kList[0].DoFindNode(&kList[targetIdx].SelfContact, searchKey)
	// // fmt.Println("KList0 KBuckets:", kList[0].table)
	// // fmt.Println("DoFindNode Result:", res_Nodes, res_Err)
	// res, _ := kList[0].DoIterativeFindNode(searchKey)
	// // res_Nodes, res_Err := kList[0].DoFindNode(&kList[2].SelfContact, searchKey)
	// // fmt.Println("DoFindNode Result:", res_Nodes, res_Err)
	// fmt.Println("IterFindNode Result:", res)
	// res = SortContact(res, searchKey)
	// fmt.Println("Result Sorted:", res)
	// if !res[0].NodeID.Equals(kList[targetIdx].SelfContact.NodeID) {
	// 	t.Error("Search result doesn't match: " + res[0].NodeID.AsString() + "!=" + kList[targetIdx].SelfContact.NodeID.AsString())
	// }
	// t.Log("TestIterativeFindNode done successfully!\n")
	// return
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
