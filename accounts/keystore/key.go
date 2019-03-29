package keystore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/herdius/herdius-core/crypto/secp256k1"
	cmn "github.com/herdius/herdius-core/libs/common"
)

// Key ...
type Key struct {
	PrivKeySP secp256k1.PrivKeySecp256k1 `json:"privKeySp"`
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
