package libkademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	mathrand "math/rand"
	"time"
	"sss"
)

type VanashingDataObject struct {
	AccessKey  int64
	Ciphertext []byte
	NumberKeys byte
	Threshold  byte
	Timeout    int
}

func GenerateRandomCryptoKey() (ret []byte) {
	for i := 0; i < 32; i++ {
		ret = append(ret, uint8(mathrand.Intn(256)))
	}
	return
}

func GenerateRandomAccessKey() (accessKey int64) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	accessKey = r.Int63()
	return
}

func CalculateSharedKeyLocations(accessKey int64, count int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func encrypt(key []byte, text []byte) (ciphertext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext = make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func decrypt(key []byte, ciphertext []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext is not long enough")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}
func extractKeysFromMap(share_map map[byte][]byte) (ret [][]byte) {
	ret = make([][]byte, 0)
  for k, v := range share_map {
		  all := append([]byte{k}, v...)
			ret = append(ret, all)
	}
	return
}

func (k *Kademlia) VanishData(data []byte, numberKeys byte, threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	key := GenerateRandomCryptoKey()
	accessKey := GenerateRandomAccessKey()
	ciphertext := encrypt(key, data)
	k.ShareKeys(numberKeys, threshold, key, accessKey)
	vdo.AccessKey = accessKey
	vdo.Ciphertext = ciphertext
	vdo.NumberKeys = numberKeys
	vdo.Threshold = threshold
	vdo.Timeout = timeoutSeconds
	go k.refresh(vdo)
	return
}

func (k *Kademlia) refresh(vdo VanashingDataObject) {
	for i := vdo.Timeout / 8; i > 0; i-- {
		select{
		    case <- time.After(time.Hour * 8):
					location_ids := CalculateSharedKeyLocations(vdo.AccessKey, (int64)(vdo.NumberKeys))
					share_map := make(map[byte][]byte)
					for _, id := range location_ids {
						val, _ := k.DoIterativeFindValue(id)
						if val != nil {
							k := val[0]
							v := val[1:]
							share_map[k] = v
						}
					}
					if len(m) < vdo.Threshold {
						fmt.Println("Not Enough Map Items!")
						return
					}
					key = sss.Combine(share_map)
					if key != nil{
						k.ShareKeys(vdo.NumberKeys, vdo.Threshold, key, vdo.AccessKey)
					}
		}
	}
}
func (k *Kademlia) ShareKeys(numberKeys byte, threshold byte, key []byte, accessKey int64) {
  share_map, err := sss.Split(numberKeys, threshold, key)
	share_keys := extractKeysFromMap(share_map)
	if err == nil {
		location_ids := CalculateSharedKeyLocations(accessKey, (int64)(numberKeys))
		for i := 0; i < (int)(numberKeys); i++ {
			  k.DoIterativeStore(location_ids[i], share_keys[i])
		}
	}
}
func (k *Kademlia) UnvanishData(vdo VanashingDataObject) (data []byte) {
	location_ids := CalculateSharedKeyLocations(vdo.AccessKey, (int64)(vdo.NumberKeys))
	share_map := make(map[byte][]byte)
	data = nil
	for _, id := range location_ids {
		val, _ := k.DoIterativeFindValue(id)
		if val != nil {
			k := val[0]
			v := val[1:]
			share_map[k] = v
		}
	}
	if len(m) < vdo.Threshold {
		fmt.Println("Not Enough Map Items!")
		return
	}
	key = sss.Combine(share_map)
	data = decrypt(key, vdo.Ciphertext)
	return
}
