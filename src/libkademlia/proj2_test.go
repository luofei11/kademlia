package libkademlia

import (
	//"bytes"
	//"net"
	//"strconv"
	"testing"
	//"time"
	"fmt"
)


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
