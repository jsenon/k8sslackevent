package bolt

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/jsenon/k8sslackevent/internal/service/cache"
	bolt "go.etcd.io/bbolt"
)

// BUCKET is the constant name of the bucket
var BUCKET = "eventssended"

// Bolt implementation
type Bolt struct {
	db *bolt.DB
}

// NewCache will return a new Bolt implementing the db.Cache interface
func NewCache() db.Cache {
	return &Bolt{}
}

// Init will create db and bucket
func (s *Bolt) Init() error {
	// Open the database.
	db, err := bolt.Open(tempfile(), 0666, nil)
	if err != nil {
		return err
	}

	s.db = db

	// Start a write transaction.
	err = s.db.Update(func(tx *bolt.Tx) error {
		// Create a bucket.
		_, err := tx.CreateBucket([]byte(BUCKET))
		return err
	})
	return err
}

func tempfile() string {
	tmpfile, err := ioutil.TempFile("", "k8sslackevent.*.db")
	if err != nil {
		log.Fatal(err)
	}
	return tmpfile.Name()
}

// CheckIfSended will check if msg is registered as already sended
func (s *Bolt) CheckIfSended(msg string) bool {
	sended := false
	// Retrieve the key again.
	if err := s.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket([]byte(BUCKET)).Get([]byte(msg))
		if value == nil {
			sended = false
		}
		sended = true
		return nil
	}); err != nil {
		return false
	}
	return sended
}

// SaveMsg will save the msg as sended in cache
func (s *Bolt) SaveMsg(msg string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(BUCKET)).Put([]byte(msg), []byte("sended"))
		return err
	})
	return err
}

// Close will delete the BoltDB
func (s *Bolt) Close() error {
	return os.Remove(s.db.Path())
}
