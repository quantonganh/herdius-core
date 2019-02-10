package keystore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	cmn "github.com/herdius/herdius-core/libs/common"
	"golang.org/x/crypto/argon2"
)

// Key ...
type Key struct {
	PrivKey ed25519.PrivKeyEd25519 `json:"privKey"`
}

// AuthKeyJSONV1 ...
type authKeyJSONV1 struct {
	Address  string     `json:"address"`
	Password string     `json:"password"`
	Crypto   CryptoJSON `json:"crypto"`
	Version  string     `json:"version"`
}

// CryptoJSON ...
type CryptoJSON struct {
	KDF       string                 `json:"kdf"`
	KDFParams map[string]interface{} `json:"kdfparams"`
}

//LoadKeyUsingPrivKey - Loads and decrypts the key from disk.
func LoadKeyUsingPrivKey(filePath string) (*Key, error) {
	if cmn.FileExists(filePath) {
		jsonBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		key := new(Key)

		err = UnmarshalJSON(jsonBytes, key)
		if err != nil {
			return nil, fmt.Errorf("Error reading Key from %v: %v", filePath, err)
		}

		return key, nil

	}
	return nil, nil
}

// LoadKeyUsingSecretKeyAndSalt - Loads and decrypts the key from disk.
// Key will be loaded based on secret, KDF details, salt details.
// Argon2 KDF is being used to derive cryptographic keys from passwords and salts.
func LoadKeyUsingSecretKeyAndSalt(filename string) (*Key, error) {
	byteValue, err := cmn.ReadFile(filename)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	var authKeyjsonv1 authKeyJSONV1

	json.Unmarshal(byteValue, &authKeyjsonv1)

	salt := authKeyjsonv1.Crypto.KDFParams["salt"].(string)
	m := authKeyjsonv1.Crypto.KDFParams["m"]
	c := authKeyjsonv1.Crypto.KDFParams["c"]

	mem := ensureUint32(m)
	cost := ensureUint32(c)
	memory := mem * cost //32*1024

	timeJSON := authKeyjsonv1.Crypto.KDFParams["time"]
	time := ensureUint32(timeJSON)

	keylenJSON := authKeyjsonv1.Crypto.KDFParams["keylen"]
	keylen := ensureUint32(keylenJSON)

	threadsJSON := authKeyjsonv1.Crypto.KDFParams["threads"]
	threads := ensureUint8(threadsJSON)

	password := authKeyjsonv1.Password
	argon2Key := argon2.Key([]byte(password), []byte(salt), time, memory, threads, keylen)
	privKey := ed25519.GenPrivKeyFromSecret(argon2Key)

	key := &Key{
		PrivKey: privKey,
	}
	return key, nil
}

// StoreKey - A key will be created, encrypted and stored in the file provided.
func StoreKey(filePath string) (*Key, error) {
	privKey := ed25519.GenPrivKey()

	key := &Key{
		PrivKey: privKey,
	}

	jsonBytes, err := MarshalJSON(key)

	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(filePath, jsonBytes, 0600)
	if err != nil {
		return nil, err
	}
	return key, nil
}

//MarshalJSON ...
func MarshalJSON(key *Key) (j []byte, err error) {
	j, err = json.Marshal(key)
	return j, err
}

// UnmarshalJSON - ...
func UnmarshalJSON(j []byte, key *Key) (err error) {
	err = json.Unmarshal(j, key)
	if err != nil {
		return err
	}
	return nil
}

func ensureUint32(x interface{}) uint32 {
	var res uint32
	if str, ok := x.(string); ok {
		res64, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return res
		}
		res = uint32(res64)
	}
	return res
}

func ensureUint8(x interface{}) uint8 {
	var res uint8
	if str, ok := x.(string); ok {
		res64, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return res
		}
		res = uint8(res64)
	}
	return res
}
