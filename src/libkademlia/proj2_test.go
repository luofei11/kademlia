package libkademlia

import (
	//"bytes"
	//"net"
	"strconv"
	"testing"
	"time"
	"fmt"
	"reflect"
)

var testPort uint16 = 3000

const testAddr = "localhost"
const divNum = 3

type KademliaList []*Kademlia

func GenerateRandomIDList(num int) (ret []ID) {
	ret = make([]ID, num)
	for i := 0; i < num; i++ {
		ret[i] = NewRandomID()
	}
	return
}

func GenerateTreeIDList(num int) (ret []ID) {
	ret = make([]ID, num)
	ret[0] = NewRandomID()
	for i := 1; i < num; i++ {
		if i > 150 {
			ret[i] = NewRandomID()
		} else {
			curID := ret[i/divNum]
			curID[i/8] = curID[i/8] ^ (1 << uint8(7-(i%8)))
			ret[i] = curID
		}
	}
	return ret
}

func GenerateTestList(num int, idList []ID) (kRet KademliaList, cRet []Contact) {
	kRet = []*Kademlia{}
	cRet = []Contact{}
	for i := 0; i < num; i++ {
		laddr := testAddr + ":" + strconv.Itoa(int(testPort))
		testPort++
		var k *Kademlia
		if idList != nil && i < len(idList) {
			k = NewKademliaWithId(laddr, idList[i])
		} //else {
			//k = NewKademliaWithId(laddr, nil)
		//}
		cRet = append(cRet, k.SelfContact)
		kRet = append(kRet, k)
	}
	return
}

func (ks KademliaList) ConnectTo(k1, k2 int) {
	ks[k1].DoPing(ks[k2].SelfContact.Host, ks[k2].SelfContact.Port)
}
func TestIterativeFindNode(t *testing.T) {
	instance1 := NewKademlia("localhost:7950")
	instance2 := NewKademlia("localhost:7951")
	instance3 := NewKademlia("localhost:7952")

	host2, port2, _ := StringToIpPort("localhost:7951")
	instance1.DoPing(host2, port2)
	host3, port3, _ := StringToIpPort("localhost:7952")
	instance2.DoPing(host3, port3)

	fmt.Println(instance1)
	fmt.Println(instance2)
	fmt.Println(instance3)

	contacts, err := instance1.DoIterativeFindNode(instance3.NodeID)
	if err != nil {
		t.Error("Error doing IterativeFindNode")
	}
	if contacts == nil || len(contacts) == 0 {
		t.Error("No contacts were found")
	}
	fmt.Println(contacts)
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
func TestIterativeFindValue(t *testing.T) {
	kNum := 120
	targetIdx := kNum - 23
	treeList := GenerateTreeIDList(kNum)
	kList, _ := GenerateTestList(kNum, treeList)
	for i := 1; i < kNum; i++ {
		kList.ConnectTo(i, i/divNum)
	}
	time.Sleep(100 * time.Millisecond)
	searchKey := kList[targetIdx].SelfContact.NodeID
	searchKey[IDBytes-1] = 0
	randValue := []byte(NewRandomID().AsString())
	kList[targetIdx/divNum].DoStore(&kList[targetIdx].SelfContact, searchKey, randValue)
	time.Sleep(3 * time.Millisecond)
	retVal, _ := kList[targetIdx].LocalFindValue(searchKey)
	if retVal == nil {
		t.Error("The target node should have the key/value pair")
		return
	}
	if string(retVal) != string(randValue) {
		t.Error("The stored value should equal to each other")
		return
	}
	res, _ := kList[0].DoIterativeFindValue(searchKey)
	if res == nil {
		t.Error("The coressponding value should be found")
		return
	}
	if string(res) != string(randValue) {
		t.Error("Search result doesn't match: " + string(res) + "!=" + string(randValue))
	}
	t.Log("TestIterativeFindValue done successfully!\n")
	return

}
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
