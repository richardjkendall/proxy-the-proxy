package main

import (
	"encoding/base64"
	"log"

	"github.com/dgraph-io/badger/v3"
)

type cache struct {
	db *badger.DB
}

func NewCache() *cache {
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatalln(`NewCache: could not open cache DB`)
	}
	return &cache{db}
}

func (c *cache) AddVal(key []byte, value []byte) error {
	log.Printf(`AddVal: adding key = %v, and value = %v`, base64.URLEncoding.EncodeToString(key), base64.URLEncoding.EncodeToString(value))
	err := c.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value)
		err := txn.SetEntry(entry)
		return err
	})
	return err
}

func (c *cache) CheckForVal(key []byte) (string, error) {
	log.Printf(`CheckForVal: checking cache for key = %v`, base64.URLEncoding.EncodeToString(key))
	var itemCopy []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			log.Printf(`CheckForVal: key not found, key = %v`, base64.URLEncoding.EncodeToString(key))
			return err
		}
		err = item.Value(func(val []byte) error {
			itemCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}

		return nil

	})
	if err != nil {
		log.Printf(`CheckForVal: got error from txn.Get, %v`, err)
		return "", err
	}
	return string(itemCopy), nil
}
