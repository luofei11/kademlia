package libkademlia

// Contains definitions for the K-Buckets.

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
)

type KBucket struct {
  var ContactList []Contact
  updateContactChan chan Contact
}

type RoutingTable [IDBits]KBucket

func NewRoutingTable() *RoutingTable{
  talbe := new(RoutingTable)
}

func (kb *KBucket) Update (c Contact) {
  kb.updateContactChan <- c
}
