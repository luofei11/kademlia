//////////////////////////////////////////////
// Test cases for some KBucket's methods
//////////////////////////////////////////////
package libkademlia

import (
	"fmt"
	"net"
	"testing"
)

func TestRemove(t *testing.T) {
	k := NewKademlia("localhost:8080")
	//selfID := k.SelfContact.NodeID
	remoteID := NewRandomID()
	c := Contact{remoteID, net.IPv4(127, 0, 0, 1), 8888}
	k.Update(c)
	if remoteContact, err := k.FindContact(remoteID); err == nil {
		if !remoteContact.NodeID.Equals(remoteID) {
			t.Error("Something wrong here.")
		}
	} else {
		t.Error("Cannot find remote's contact.")
	}
	selfkbindex := k.FindBucket(remoteID)
	fmt.Printf("%s",selfkbindex)
	kb := &k.table[selfkbindex]
	contains_1, i := kb.FindContactInKBucket(c)
	if !contains_1 {
		t.Error("Can't find remote's contact in self's contact list.")
	} else {
		kb.Remove(i)
	}
	if _, err := k.FindContact(remoteID); err == nil {
		t.Error("This ID has been removed.")
	}
}

func TestMoveToTail(t *testing.T) {
	k := NewKademlia("localhost:5000")
	remoteID := NewRandomID()
	c := Contact{remoteID, net.IPv4(127, 0, 0, 1), 5050}
	k.Update(c)
	if remoteContact, err := k.FindContact(remoteID); err == nil {
		if !remoteContact.NodeID.Equals(remoteID) {
			t.Error("Something wrong here.")
		}
	} else {
		t.Error("Cannot find remote's contact.")
	}
	selfkbindex := k.FindBucket(remoteID)
	fmt.Printf("%s",selfkbindex)
	kb := &k.table[selfkbindex]
	contains, i := kb.FindContactInKBucket(c)
	if !contains {
		t.Error("Can't find remote's contact in self's contact list.")
	} else {
		kb.MoveToTail(i)
	}
	length := len(*kb)
	contains_2, tail1 := kb.FindContactInKBucket(c)
	if !contains_2 {
		t.Error("Can't find remote's contact in self's contact list.")
	} else {
		if length - 1 != tail1 {
			t.Error("Can't move to tail.")
		} 
	}
}

func TestAddToTail(t *testing.T) {
	k := NewKademlia("localhost:4000")
	remoteID := NewRandomID()
	c := Contact{remoteID, net.IPv4(127, 0, 0, 1), 4004}
	k.Update(c)
	if remoteContact, err := k.FindContact(remoteID); err == nil {
		if !remoteContact.NodeID.Equals(remoteID) {
			t.Error("Something wrong here.")
		}
	} else {
		t.Error("Cannot find remote's contact.")
	}
	selfkbindex := k.FindBucket(remoteID)
	fmt.Printf("%s",selfkbindex)
	kb := &k.table[selfkbindex]
	contains, i := kb.FindContactInKBucket(c)
	if !contains {
		t.Error("Can't find remote's contact in self's contact list.")
	} else {
		kb.Remove(i)
		kb.AddToTail(c)
	}
	length := len(*kb)
	contains_3, tail2 := kb.FindContactInKBucket(c)
	if !contains_3 {
		t.Error("Can't find remote's contact in self's contact list.")
	} else {
		if length - 1 != tail2 {
			t.Error("Can't add to tail.")
		} 
	}
}