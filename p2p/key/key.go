package key

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/herdius/herdius-core/crypto"
	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	cmn "github.com/herdius/herdius-core/libs/common"
)

// ID is a hex-encoded crypto.Address
type ID string

// IDByteLength is the length of a crypto.Address. Currently only 20.
const IDByteLength = crypto.AddressSize

// NodeKey is the persistent peer key.
// It contains the nodes private key for authentication.
type NodeKey struct {
	PrivKey crypto.PrivKey `json:"priv_key"` // our priv key
}

// ID returns the peer's canonical ID - the hash of its public key.
func (nodeKey *NodeKey) ID() ID {
	return PubKeyToID(nodeKey.PubKey())
}

// PubKey returns the peer's PubKey
func (nodeKey *NodeKey) PubKey() crypto.PubKey {
	return nodeKey.PrivKey.PubKey()
}

// PubKeyToID returns the ID corresponding to the given PubKey.
// It's the hex-encoding of the pubKey.Address().
func PubKeyToID(pubKey crypto.PubKey) ID {
	return ID(hex.EncodeToString(pubKey.Address()))
}

// LoadOrGenNodeKey attempts to load the NodeKey from the given filePath.
// If the file does not exist, it generates and saves a new NodeKey.
func LoadOrGenNodeKey(filePath string) (*NodeKey, error) {
	if cmn.FileExists(filePath) {
		nodeKey, err := LoadNodeKey(filePath)
		if err != nil {
			return nil, err
		}
		return nodeKey, nil
	}
	return genNodeKey(filePath)
}

// LoadNodeKey ...
func LoadNodeKey(filePath string) (*NodeKey, error) {
	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	nodeKey := new(NodeKey)
	err = cdc.UnmarshalJSON(jsonBytes, nodeKey)
	if err != nil {
		return nil, fmt.Errorf("Error reading NodeKey from %v: %v", filePath, err)
	}
	return nodeKey, nil
}

func genNodeKey(filePath string) (*NodeKey, error) {
	privKey := ed25519.GenPrivKey()
	nodeKey := &NodeKey{
		PrivKey: privKey,
	}

	jsonBytes, err := cdc.MarshalJSON(nodeKey)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filePath, jsonBytes, 0600)
	if err != nil {
		return nil, err
	}
	return nodeKey, nil
}
