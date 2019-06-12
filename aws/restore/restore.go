package restore

import (
	"fmt"

	"github.com/herdius/herdius-core/config"
)

type RestorerI interface {
	Restore() error
	testCompleteChainRemote() (bool, error)
	clearOldChain() error
	downloadChain() error
	replayChain() error
}

type Restorer struct {
	statePath       string
	chainPath       string
	s3bucket        string
	heightToRestore int
	prefix          string
}

func NewRestorer(env string, height int) RestorerI {
	detail := config.GetConfiguration(env)
	return Restorer{
		statePath:       detail.StateDBPath,
		chainPath:       detail.ChainDBPath,
		s3bucket:        detail.S3Bucket,
		heightToRestore: height,
	}
}

// Restore retrieves and procceses an entire blockchain stored in S3
// into the Supervisor's local blockchain and statedb
func (r Restorer) Restore() error {
	succ, err := r.testCompleteChainRemote()
	if err != nil {
		return fmt.Errorf("could not restore chain from backup: %v", err)
	}
	if !succ {
		return fmt.Errorf("could not restore chain from backup, specified chain in S3 is invalid")
	}

	err = r.clearOldChain()
	if err != nil {
		return fmt.Errorf("restore failed while trying to clean old chain: %v", err)
	}

	err = r.downloadChain()
	if err != nil {
		return fmt.Errorf("restore failed while trying to download backed up chain: %v", err)
	}

	err = r.replayChain()
	if err != nil {
		return fmt.Errorf("restore failed while trying to replay chain: %v", err)
	}

	return fmt.Errorf("unable to restore chain entirely, but reached end of Remote()")
}

func (r Restorer) testCompleteChainRemote() (bool, error) {
	return true, nil
}

func (r Restorer) clearOldChain() error {
	return nil
}

func (r Restorer) downloadChain() error {}

func (r Restorer) replayChain() error {}
