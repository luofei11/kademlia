package libkademlia

// Contains definitions for the RoutingTable and K-Bucket.

type KBucket []Contact

type RoutingTable [IDBits]KBucket

func (table *RoutingTable) Initialize() {
  for i := 0; i < IDBits; i++ {
    table[i] = make([]Contact, 0, k)
  }
}

func (kb *KBucket) FindContactInKBucket (c Contact) (bool, int) {
  for i := 0; i < len(kb); i++ {
		temp := kb[i]
		if temp.NodeID.Equals(c.NodeID) {
      return true, i
		}
	}
  return false, -1
}

func (kb *KBucket) Remove (i int) {
	kb = append(kb[:i], kb[i+1:]...)
}

func (kb *KBucket) AddToTail (c Contact) {
  kb = append(kb, c)
}

func (kb *KBucket) MoveToTail (i int) {
  c := kb[i]
  kb.Remove(i)
  kb.AddToTail(c)
}
