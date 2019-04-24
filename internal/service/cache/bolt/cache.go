package bolt

import (
	"crypto/sha1"
	"io/ioutil"
	"log"
	"os"

	db "github.com/jsenon/k8sslackevent/internal/service/cache"
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
		_, errBucket := tx.CreateBucket([]byte(BUCKET))
		return errBucket
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
	key := sha1.Sum([]byte(msg))
	sended := false
	// Retrieve the key again.
	if err := s.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket([]byte(BUCKET)).Get(key[:])
		if value == nil {
			sended = false
			return nil
		}
		sended = true
		return nil
	}); err != nil {
		log.Println(err.Error())
		return false
	}
	return sended
}

// SaveMsg will save the msg as sended in cache
func (s *Bolt) SaveMsg(msg string) error {
	key := sha1.Sum([]byte(msg))
	err := s.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(BUCKET)).Put(key[:], []byte("msg"))
		return err
	})
	return err
}

// Close will delete the BoltDB
func (s *Bolt) Close() error {
	if err := s.db.Close(); err != nil {
		log.Println(err.Error())
	}
	return os.Remove(s.db.Path())
}
