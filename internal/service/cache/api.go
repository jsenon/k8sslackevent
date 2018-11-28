package db

// Cache describe interface to save k8s event pushed on slack
type Cache interface {
	Init() error
	CheckIfSended(msg string) bool
	SaveMsg(msg string) error
	Close() error
}
