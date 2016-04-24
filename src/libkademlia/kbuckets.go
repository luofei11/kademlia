package libkademlia

// Contains definitions for the K-Buckets.

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
)

type KBucket []Contact

type RoutingTable [IDBits]KBucket

func (table *RoutingTable) Initialize() {
  for i := 0; i < IDBits; i++ {
    table[i] = make([]Contact, 0, k)
  }
}
