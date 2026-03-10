package core

import (
	"encoding/json"
	"fmt"
	"log"

	"go.etcd.io/bbolt"
)

// Storage handles local persistence using BoltDB (bbolt)
type Storage struct {
	db *bbolt.DB
}

// NewStorage initializes a new BoltDB storage engine
func NewStorage(dbPath string) (*Storage, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt database: %w", err)
	}
	return &Storage{db: db}, nil
}

// Close gracefully closes the DB
func (s *Storage) Close() error {
	return s.db.Close()
}

// EnsureCollection creates a bucket for the collection if it doesn't exist
func (s *Storage) EnsureCollection(colName string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(colName))
		return err
	})
}

// UpsertPoints saves or updates a batch of points to disk
func (s *Storage) UpsertPoints(colName string, points []PointStruct) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(colName))
		if b == nil {
			return fmt.Errorf("collection bucket %s not found", colName)
		}

		for _, p := range points {
			data, err := json.Marshal(p)
			if err != nil {
				return fmt.Errorf("failed to marshal point %s: %w", p.ID, err)
			}
			if err := b.Put([]byte(p.ID), data); err != nil {
				return fmt.Errorf("failed to write point %s to bbolt: %w", p.ID, err)
			}
		}
		return nil
	})
}

// LoadCollection loads all points for a collection from disk into memory
func (s *Storage) LoadCollection(colName string) (map[string]*PointStruct, error) {
	points := make(map[string]*PointStruct)

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(colName))
		if b == nil {
			log.Printf("Bucket %s doesn't exist yet, starting empty", colName)
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var p PointStruct
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			points[string(k)] = &p
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load points from bbolt: %w", err)
	}
	return points, nil
}

// DeletePoints deletes a batch of points from disk
func (s *Storage) DeletePoints(colName string, ids []string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(colName))
		if b == nil {
			return nil // Bucket doesn't exist, nothing to delete
		}

		for _, id := range ids {
			if err := b.Delete([]byte(id)); err != nil {
				return fmt.Errorf("failed to delete point %s from bbolt: %w", id, err)
			}
		}
		return nil
	})
}
