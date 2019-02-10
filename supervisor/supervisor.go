package supervisor

// ID of Supervisor
type ID struct {
	// public_key of the peer (we no longer use the public key as the peer ID, but use it to verify messages)
	PublicKey []byte `json:"public_key,omitempty"`
	// address is the network address of the peer
	Address string `json:"address,omitempty"`
	// ID is the computed hash of the public key
	ID []byte `json:"id,omitempty"`
}
