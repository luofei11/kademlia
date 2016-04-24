package libkademlia

// Contains definitions for the K-Buckets.

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
)

type KBucket struct {
  ContactList []Contact
  updateContactChan chan Contact
}

type RoutingTable [IDBits]KBucket

func (table *RoutingTable) Initialize() {
  for i := 0; i < IDBits; i++ {
    table[i].ContactList = make([]Contact, 0, k)
    table[i].updateContactChan = make(chan Contact)
  }
}

func (kb *KBucket) Update (c Contact) {
  kb.updateContactChan <- c
}
