package sync

type Syncer interface {
	GetExtBalance()
	Update()
}
