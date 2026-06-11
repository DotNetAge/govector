package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"unicode/utf8"

	pb "github.com/DotNetAge/govector/core/proto"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

// Storage handles local persistence using BoltDB (bbolt).
// It provides durable storage for vector collections and their points.
type Storage struct {
	db        *bbolt.DB
	closed    bool
	quantizer Quantizer
	useQuant  bool
}

func toProtoPoint(p *PointStruct) *pb.PointStruct {
	pbPoint := &pb.PointStruct{
		Id:     p.ID,
		Vector: p.Vector,
	}

	if p.Payload != nil {
		pbPoint.Payload = make(map[string]*pb.Value)
		for k, v := range p.Payload {
			pbVal := &pb.Value{}
			switch val := v.(type) {
			case string:
				if !utf8.ValidString(val) {
					val = string(bytes.ToValidUTF8([]byte(val), nil)) // sanitize: replace invalid UTF-8 sequences
				}
				pbVal.Value = &pb.Value_StringValue{StringValue: val}
			case int:
				pbVal.Value = &pb.Value_IntValue{IntValue: int64(val)}
			case int64:
				pbVal.Value = &pb.Value_IntValue{IntValue: val}
			case float32:
				pbVal.Value = &pb.Value_DoubleValue{DoubleValue: float64(val)}
			case float64:
				pbVal.Value = &pb.Value_DoubleValue{DoubleValue: val}
			case bool:
				pbVal.Value = &pb.Value_BoolValue{BoolValue: val}
			case []byte:
				pbVal.Value = &pb.Value_BytesValue{BytesValue: val}
			default:
				// For complex types (map, slice, etc.), JSON-serialize and store as bytes
				if jsonBytes, err := json.Marshal(val); err == nil {
					pbVal.Value = &pb.Value_BytesValue{BytesValue: jsonBytes}
				}
				// If JSON marshal fails, skip the value
			}
			pbPoint.Payload[k] = pbVal
		}
	}

	return pbPoint
}

func fromProtoPoint(pbPoint *pb.PointStruct) *PointStruct {
	p := &PointStruct{
		ID:     pbPoint.Id,
		Vector: pbPoint.Vector,
	}

	if pbPoint.Payload != nil {
		p.Payload = make(Payload)
		for k, v := range pbPoint.Payload {
			if v.Value == nil {
				continue
			}
			switch val := v.Value.(type) {
			case *pb.Value_StringValue:
				p.Payload[k] = val.StringValue
			case *pb.Value_IntValue:
				p.Payload[k] = val.IntValue
			case *pb.Value_DoubleValue:
				p.Payload[k] = val.DoubleValue
			case *pb.Value_BoolValue:
				p.Payload[k] = val.BoolValue
			case *pb.Value_BytesValue:
				// Try to deserialize as JSON for complex nested objects (map, slice, etc.)
				// If it's not valid JSON, return raw bytes
				var obj any
				if err := json.Unmarshal(val.BytesValue, &obj); err == nil {
					p.Payload[k] = obj
				} else {
					p.Payload[k] = val.BytesValue
				}
			}
		}
	}

	return p
}

// NewStorage initializes a new BoltDB storage engine at the specified path.
// The database file will be created if it doesn't exist.
// Returns an error if the database cannot be opened.
func NewStorage(dbPath string) (*Storage, error) {
	return NewStorageWithQuantization(dbPath, false, nil, false)
}

// NewStorageWithQuantization initializes a new BoltDB storage engine with optional vector quantization.
// If useQuant is true, vectors will be compressed using the provided quantizer.
// When readOnly is true, opens the database in read-only mode (shared lock), allowing
// coexistence with another process that holds the write lock (e.g. a running Daemon).
// Returns an error if the database cannot be opened.
func NewStorageWithQuantization(dbPath string, useQuant bool, quantizer Quantizer, readOnly bool) (*Storage, error) {
	opts := &bbolt.Options{ReadOnly: readOnly}
	mode := os.FileMode(0600)
	if readOnly {
		mode = 0400
		// When opening as read-only, another process (e.g. a running Daemon)
		// may hold the exclusive write lock (LOCK_EX) on this file. Without a
		// timeout the read lock attempt (LOCK_SH) blocks forever. A short
		// timeout ensures we fail fast with an actionable error instead of
		// hanging the entire process indefinitely.
		opts.Timeout = 3 * time.Second
	} else {
		// Write mode also needs a timeout: if another process (e.g. a stale daemon,
		// or a leftover test process) already holds the exclusive lock, bbolt.Open()
		// will block forever waiting for it. A reasonable timeout prevents the
		// entire daemon from hanging at startup.
		opts.Timeout = 5 * time.Second
	}
	db, err := bbolt.Open(dbPath, mode, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt database: %w", err)
	}

	if useQuant && quantizer == nil {
		quantizer = NewSQ8Quantizer()
	}

	return &Storage{
		db:        db,
		closed:    false,
		quantizer: quantizer,
		useQuant:  useQuant,
	}, nil
}

// Close gracefully closes the database connection.
// It is important to call this when done to ensure all data is flushed to disk.
// This method is idempotent and can be called multiple times.
func (s *Storage) Close() error {
	if s.closed {
		return nil
	}

	err := s.db.Close()
	if err == nil {
		s.closed = true
	}
	return err
}

// EnsureCollection creates a bucket for the collection if it doesn't already exist.
// Each collection is stored as a separate bucket in BoltDB.
func (s *Storage) EnsureCollection(colName string) error {
	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(colName))
		return err
	})
}

// UpsertPoints saves or updates a batch of points to disk.
// Points are serialized as Protocol Buffers and stored with their ID as the key.
// If quantization is enabled, vectors are compressed before storage.
// Returns an error if the collection doesn't exist or if serialization fails.
func (s *Storage) UpsertPoints(colName string, points []PointStruct) error {
	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(colName))
		if b == nil {
			return fmt.Errorf("collection bucket %s not found", colName)
		}

		for i := range points {
			p := &points[i]

			// Quantize vector if enabled
			if s.useQuant && s.quantizer != nil {
				// Create a copy to avoid modifying the original
				quantizedPoint := *p
				quantizedPoint.Vector = nil // Clear vector for storage

				// Store quantized vector in payload
				if quantizedPoint.Payload == nil {
					quantizedPoint.Payload = make(Payload)
				}
				quantizedPoint.Payload["__quantized_vector"] = s.quantizer.Quantize(p.Vector)

				p = &quantizedPoint
			}

			data, err := proto.Marshal(toProtoPoint(p))
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

// LoadCollection loads all points for a collection from disk into memory.
// Returns a map of point IDs to PointStruct pointers.
// If quantization is enabled, compressed vectors are decompressed during loading.
// If the collection doesn't exist, returns an empty map without error.
func (s *Storage) LoadCollection(colName string) (map[string]*PointStruct, error) {
	if s.closed {
		return nil, fmt.Errorf("storage is closed")
	}

	points := make(map[string]*PointStruct)

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(colName))
		if b == nil {
			log.Printf("Bucket %s doesn't exist yet, starting empty", colName)
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var pbPoint pb.PointStruct
			if err := proto.Unmarshal(v, &pbPoint); err != nil {
				return err
			}

			p := fromProtoPoint(&pbPoint)

			// Dequantize vector if enabled and quantized vector exists
			if s.useQuant && s.quantizer != nil {
				if quantizedData, ok := p.Payload["__quantized_vector"]; ok {
					if data, ok := quantizedData.([]byte); ok {
						p.Vector = s.quantizer.Dequantize(data)
						// Remove quantized vector from payload to save memory
						delete(p.Payload, "__quantized_vector")
					}
				}
			}

			points[string(k)] = p
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load points from bbolt: %w", err)
	}
	return points, nil
}

// ListCollections returns all collection names (bucket names) in the storage.
// It excludes the internal metadata bucket.
func (s *Storage) ListCollections() ([]string, error) {
	if s.closed {
		return nil, fmt.Errorf("storage is closed")
	}

	var collections []string
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			colName := string(name)
			if colName != "__collections_meta__" {
				collections = append(collections, colName)
			}
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	return collections, nil
}

// SaveCollectionMeta persists collection metadata to a special bucket.
func (s *Storage) SaveCollectionMeta(name string, meta CollectionMeta) error {
	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__collections_meta__"))
		if err != nil {
			return err
		}

		data, err := json.Marshal(meta)
		if err != nil {
			return err
		}

		return b.Put([]byte(name), data)
	})
}

// LoadCollectionMeta retrieves collection metadata from the special bucket.
func (s *Storage) LoadCollectionMeta(name string) (*CollectionMeta, error) {
	if s.closed {
		return nil, fmt.Errorf("storage is closed")
	}

	var meta CollectionMeta
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("__collections_meta__"))
		if b == nil {
			return fmt.Errorf("metadata bucket not found")
		}

		data := b.Get([]byte(name))
		if data == nil {
			return fmt.Errorf("metadata for collection %s not found", name)
		}

		return json.Unmarshal(data, &meta)
	})

	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// ListCollectionMetas returns all collection metadata stored in the database.
func (s *Storage) ListCollectionMetas() ([]CollectionMeta, error) {
	if s.closed {
		return nil, fmt.Errorf("storage is closed")
	}

	var metas []CollectionMeta
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("__collections_meta__"))
		if b == nil {
			return nil // No metadata stored yet
		}

		return b.ForEach(func(k, v []byte) error {
			var meta CollectionMeta
			if err := json.Unmarshal(v, &meta); err != nil {
				return err
			}
			metas = append(metas, meta)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list collection metas: %w", err)
	}
	return metas, nil
}

// DeletePoints deletes a batch of points from disk by their IDs.
// If a point ID doesn't exist, it is silently skipped.
// Returns an error if the collection doesn't exist.
func (s *Storage) DeletePoints(colName string, ids []string) error {
	if s.closed {
		return fmt.Errorf("storage is closed")
	}

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

// DropCollection removes a collection and its metadata from the storage.
func (s *Storage) DropCollection(name string) error {
	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		// Delete collection bucket
		if err := tx.DeleteBucket([]byte(name)); err != nil && err != bbolt.ErrBucketNotFound {
			return fmt.Errorf("failed to delete bucket: %w", err)
		}

		// Delete metadata
		metaBucket := tx.Bucket([]byte("__collections_meta__"))
		if metaBucket != nil {
			if err := metaBucket.Delete([]byte(name)); err != nil {
				return fmt.Errorf("failed to delete metadata: %w", err)
			}
		}

		return nil
	})
}
