package entity

import (
	"log"

	"github.com/boltdb/bolt"
)

var BDB *bolt.DB

func OpenBoltDB() {
	var err error
	BDB, err = bolt.Open("md.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	BDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("meta"))
		if err != nil {
			return err
		}
		return nil
	})

	BDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("delete"))
		if err != nil {
			return err
		}
		return nil
	})
}

func BoltSave(bkt []byte, key []byte, val []byte) error {
	return BDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bkt)
		err := b.Put(key, val)
		return err
	})
}

func BoltDelete(bkt []byte, key []byte) error {
	return BDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bkt)
		err := b.Delete(key)
		log.Println(err)
		return err
	})
}

func BoltList(bkt []byte, pageSize int) (l [][]byte) {
	if pageSize <= 0 {
		pageSize = 1000
	}
	BDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bkt)
		c := b.Cursor()
		idx := 0
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if idx >= pageSize {
				break
			}
			l = append(l, k)
			idx++
		}

		return nil
	})
	return l
}
