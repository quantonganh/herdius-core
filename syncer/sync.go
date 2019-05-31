package sync

type Syncer interface {
	GetExtBalance() error
	Update()
}
