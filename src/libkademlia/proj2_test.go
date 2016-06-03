package libkademlia

import (
	//"bytes"
	//"net"
	"strconv"
	"testing"
	"sort"
	//"time"
	"fmt"
	//"reflect"
	//"container/heap"
)

func GenerateTreeKademlia(num_treenode int, start_port int) []*Kademlia {
	ResultList := make([]*Kademlia, 0, num_treenode)
	root_address := "localhost:" + strconv.Itoa(start_port)
	root_kademlia := NewKademlia(root_address)
	ResultList = append(ResultList, root_kademlia)
	for i := 1; i < num_treenode; i++ {
		leaf_address := "localhost:" + strconv.Itoa(start_port + i)
		leaf_NodeID := ResultList[i / 3].NodeID
		leaf_NodeID[i/8] = leaf_NodeID[i/8] ^ (1 << uint8(7-(i%8)))
		leaf_kademlia := NewKademliaWithId(leaf_address, leaf_NodeID)
		ResultList = append(ResultList, leaf_kademlia)
		father_address := "localhost:" + strconv.Itoa(start_port + i / 3)
		host_number, port_number, _ := StringToIpPort(father_address)
		ResultList[i].DoPing(host_number, port_number)
	}
	return ResultList
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
	num_treenode := 27
	target := 16
	tree_kademlia := GenerateTreeKademlia(num_treenode, 8050)
	searchKey := tree_kademlia[target].NodeID
	searchValue := []byte("hello world!")
	//fmt.Println("Search value is: ", searchValue)
	tree_kademlia[target / 3].DoStore(&tree_kademlia[target].SelfContact, searchKey, searchValue)
	v, err := tree_kademlia[target].LocalFindValue(searchKey)
	if err != nil {
		t.Error("DoStore error!")
		return
	}
	if string(v) != string(searchValue) {
		t.Error("Value doesn't match!")
		return
	}
	//tree_kademlia[0].DoFindNode(&tree_kademlia[target].SelfContact, searchKey)

	fmt.Println("Last Node:", *tree_kademlia[num_treenode - 1])

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
func TestIterativeFindValueSimple(t *testing.T) {
	instance1 := NewKademlia("localhost:8939")
	instance2 := NewKademlia("localhost:8940")
	instance3 := NewKademlia("localhost:8941")

	host2, port2, _ := StringToIpPort("localhost:8940")
	instance1.DoPing(host2, port2)
	host3, port3, _ := StringToIpPort("localhost:8941")
	instance2.DoPing(host3, port3)
	searchKey := instance3.SelfContact.NodeID
	searchKey[IDBytes - 1] = 0
	searchValue := []byte("helloworld!")
	instance1.DoStore(&instance3.SelfContact, searchKey, searchValue)
	v, _ := instance2.DoIterativeFindValue(searchKey)
	if v == nil{
		t.Error("Simple IterFindValue failed")
	}else{
		t.Log("Simple IterFindValue succeed!")
	}
	return
}


func TestIterativeStore(t *testing.T) {
	num_treenode := 27
	target := 24
	tree_kademlia := GenerateTreeKademlia(num_treenode, 8100)
	searchKey := tree_kademlia[target].NodeID
	searchValue := []byte("hello world!")
	//fmt.Println("Search value is: ", searchValue)
	_, err := tree_kademlia[0].DoIterativeStore(searchKey, searchValue)
  if err != nil {
    t.Error("IterativeStore returns error!")
		return
	}
	v, err := tree_kademlia[target].LocalFindValue(searchKey)
	if err != nil {
		t.Error("didn't find value!")
		return
	}
	if string(v) != string(searchValue) {
		t.Error("Value doesn't match!")
		return
	}
	return
}
