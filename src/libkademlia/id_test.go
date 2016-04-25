package libkademlia

import (
  "strings"
	//"bytes"
	//"net"
	//"strconv"
	"testing"
	//"time"
)

func TestXor(t *testing.T){
    var id_instance1 ID
    var id_instance2 ID
    var result1_should_be ID
    for i := 0; i < IDBytes; i++{
        id_instance1[i] = uint8(0)
        id_instance2[i] = uint8(1)
        result1_should_be[i] = uint8(1)
    }
    test1_result := id_instance1.Xor(id_instance2)
    test2_result := id_instance2.Xor(id_instance2)
    if strings.Compare(test1_result.AsString(), result1_should_be.AsString()) != 0{
        t.Error("Xor wrong!")
    }
    if strings.Compare(test2_result.AsString(), id_instance1.AsString()) != 0{
        t.Error("Xor wrong!")
    }
    return
}

func TestCopy(t *testing.T){
    for i := 0; i < 100; i++{
        old_ID := NewRandomID()
        new_ID := CopyID(old_ID)
        if strings.Compare(old_ID.AsString(), new_ID.AsString()) != 0{
            t.Error("ID Copy wrong!")
        }
    }
    return
}

func TestIDFromString(t *testing.T){
    for i := 0; i < 100; i++{
        old_ID := NewRandomID()
        old_IDString := old_ID.AsString()
        original_ID, err := IDFromString(old_IDString)
        if err != nil{
            t.Error("ID From String wrong!")
        }
        if !old_ID.Equals(original_ID){
            t.Error("ID From String wrong!")
        }
    }
    return
}

func TestCompare(t *testing.T){
  var id_instance1 ID
  var id_instance2 ID
  for i := 0; i < IDBytes; i++{
      id_instance1[i] = uint8(0)
      id_instance2[i] = uint8(1)
  }
  if id_instance1.Compare(id_instance1) != 0{
      t.Error("ID Compare wrong!")
  }
  id_instance3 := NewRandomID()
  id_instance4 := NewRandomID()
  if id_instance3.Compare(id_instance4) == 0{
      t.Error("ID From String wrong!")
  }
  return
}
