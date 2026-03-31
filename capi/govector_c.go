// Package main provides CGO bindings for GoVector, enabling cross-language integration.
//
// This package serves as an adapter layer between GoVector's pure Go core
// and C-based languages (Python, C++, Java, etc.) through SWIG.
//
// Architecture:
//   - Pure Go core: github.com/DotNetAge/govector/core (no CGO dependencies)
//   - CGO adapter: this package (govector/capi)
//   - One-way dependency: capi -> core (no circular dependencies)
package main

/*
#include "govector_c.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"

	"github.com/DotNetAge/govector/core"
)

// ============================================================================
// Global Registry - Manages handles for C code
// ============================================================================

var (
	storageRegistry    = make(map[uint64]*core.Storage)
	collectionRegistry = make(map[uint64]*core.Collection)
	registryMu         sync.RWMutex

	nextStorageID    uint64 = 1
	nextCollectionID uint64 = 1
)

// ============================================================================
// Type Conversions
// ============================================================================

// toDistance converts C DistanceType to Go core.Distance
func toDistance(metric C.DistanceType) core.Distance {
	switch metric {
	case C.DISTANCE_COSINE:
		return core.Cosine
	case C.DISTANCE_EUCLID:
		return core.Euclid
	case C.DISTANCE_DOT:
		return core.Dot
	default:
		return core.Cosine
	}
}

// toHNSWParams converts C HNSWParams to Go core.HNSWParams
func toHNSWParams(params C.HNSWParams) core.HNSWParams {
	return core.HNSWParams{
		M:              int(params.m),
		EfConstruction: int(params.ef_construction),
		EfSearch:       int(params.ef_search),
		K:              int(params.k),
	}
}

// pointStructFromC converts C PointStruct to Go core.PointStruct
func pointStructFromC(cPoint *C.PointStruct) core.PointStruct {
	point := core.PointStruct{
		ID:      C.GoString(cPoint.id),
		Version: uint64(cPoint.version),
	}

	// Convert vector
	if cPoint.vector != nil && cPoint.vector_dim > 0 {
		vectorLen := int(cPoint.vector_dim)
		point.Vector = make([]float32, vectorLen)
		cVector := (*[1 << 30]C.float)(unsafe.Pointer(cPoint.vector))[:vectorLen:vectorLen]
		for i := 0; i < vectorLen; i++ {
			point.Vector[i] = float32(cVector[i])
		}
	}

	// Convert payload from JSON
	if cPoint.payload_json != nil {
		jsonStr := C.GoString(cPoint.payload_json)
		json.Unmarshal([]byte(jsonStr), &point.Payload)
	}

	return point
}

// scoredPointToC converts Go core.ScoredPoint to C ScoredPoint
func scoredPointToC(goPoint core.ScoredPoint) C.ScoredPoint {
	cPoint := C.ScoredPoint{
		id:      C.CString(goPoint.ID),
		version: C.uint64_t(goPoint.Version),
		score:   C.float(goPoint.Score),
	}

	// Convert payload to JSON
	if goPoint.Payload != nil {
		jsonData, _ := json.Marshal(goPoint.Payload)
		cPoint.payload_json = C.CString(string(jsonData))
	}

	return cPoint
}

// newError creates a new ErrorInfo with the given code and message
func newError(code C.int, msg string) C.ErrorInfo {
	return C.ErrorInfo{
		code:    code,
		message: C.CString(msg),
	}
}

// ============================================================================
// Storage Engine API Implementation
// ============================================================================

func govector_storage_new(dbPath *C.char, errInfo *C.ErrorInfo) C.StorageHandle {
	path := C.GoString(dbPath)

	store, err := core.NewStorage(path)
	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_STORAGE_FAILURE, err.Error())
		}
		return nil
	}

	registryMu.Lock()
	handle := nextStorageID
	nextStorageID++
	storageRegistry[handle] = store
	registryMu.Unlock()

	return C.StorageHandle(uintptr(handle))
}

func govector_storage_close(handle C.StorageHandle) C.int {
	registryMu.Lock()
	defer registryMu.Unlock()

	store, exists := storageRegistry[uint64(uintptr(handle))]
	if !exists {
		return C.GOVECTOR_ERROR_NOT_FOUND
	}

	if err := store.Close(); err != nil {
		return C.GOVECTOR_ERROR_STORAGE_FAILURE
	}

	delete(storageRegistry, uint64(uintptr(handle)))
	return C.GOVECTOR_OK
}

func govector_storage_free(handle C.StorageHandle) {
	registryMu.Lock()
	defer registryMu.Unlock()

	delete(storageRegistry, uint64(uintptr(handle)))
}

// ============================================================================
// Collection Management API Implementation
// ============================================================================

func govector_collection_create(
	name *C.char,
	vectorDim C.int,
	metric C.DistanceType,
	useHnsw C.bool,
	hnswParams C.HNSWParams,
	storage C.StorageHandle,
	errInfo *C.ErrorInfo,
) C.CollectionHandle {
	colName := C.GoString(name)

	var store *core.Storage
	if storage != nil {
		registryMu.RLock()
		store = storageRegistry[uint64(uintptr(storage))]
		registryMu.RUnlock()
		if store == nil {
			if errInfo != nil {
				*errInfo = newError(C.GOVECTOR_ERROR_NOT_FOUND, "storage not found")
			}
			return nil
		}
	}

	params := toHNSWParams(hnswParams)
	col, err := core.NewCollectionWithParams(
		colName,
		int(vectorDim),
		toDistance(metric),
		store,
		bool(useHnsw),
		params,
	)

	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_GENERAL, err.Error())
		}
		return nil
	}

	registryMu.Lock()
	handle := nextCollectionID
	nextCollectionID++
	collectionRegistry[handle] = col
	registryMu.Unlock()

	return C.CollectionHandle(uintptr(handle))
}

func govector_collection_load(
	name *C.char,
	storage C.StorageHandle,
	errInfo *C.ErrorInfo,
) C.CollectionHandle {
	colName := C.GoString(name)

	var store *core.Storage
	if storage != nil {
		registryMu.RLock()
		store = storageRegistry[uint64(uintptr(storage))]
		registryMu.RUnlock()
		if store == nil {
			if errInfo != nil {
				*errInfo = newError(C.GOVECTOR_ERROR_NOT_FOUND, "storage not found")
			}
			return nil
		}
	}

	meta, err := store.LoadCollectionMeta(colName)
	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_NOT_FOUND, fmt.Sprintf("collection %s not found", colName))
		}
		return nil
	}

	col, err := core.NewCollectionWithParams(
		meta.Name,
		meta.VectorLen,
		meta.Metric,
		store,
		meta.UseHNSW,
		meta.HNSWParams,
	)

	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_GENERAL, err.Error())
		}
		return nil
	}

	registryMu.Lock()
	handle := nextCollectionID
	nextCollectionID++
	collectionRegistry[handle] = col
	registryMu.Unlock()

	return C.CollectionHandle(uintptr(handle))
}

func govector_collection_drop(
	name *C.char,
	storage C.StorageHandle,
	errInfo *C.ErrorInfo,
) C.int {
	colName := C.GoString(name)

	var store *core.Storage
	if storage != nil {
		registryMu.RLock()
		store = storageRegistry[uint64(uintptr(storage))]
		registryMu.RUnlock()
		if store == nil {
			if errInfo != nil {
				*errInfo = newError(C.GOVECTOR_ERROR_NOT_FOUND, "storage not found")
			}
			return C.GOVECTOR_ERROR_NOT_FOUND
		}
	}

	if err := store.DropCollection(colName); err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_GENERAL, err.Error())
		}
		return C.GOVECTOR_ERROR_GENERAL
	}

	return C.GOVECTOR_OK
}

func govector_collection_free(handle C.CollectionHandle) {
	registryMu.Lock()
	defer registryMu.Unlock()

	delete(collectionRegistry, uint64(uintptr(handle)))
}

// ============================================================================
// Point Operations API Implementation
// ============================================================================

func govector_collection_upsert(
	handle C.CollectionHandle,
	points *C.PointStruct,
	count C.int,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	col := collectionRegistry[uint64(uintptr(handle))]
	registryMu.RUnlock()

	if col == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_NOT_FOUND, "collection not found")
		}
		return C.GOVECTOR_ERROR_NOT_FOUND
	}

	// Convert points array
	goPoints := make([]core.PointStruct, int(count))
	cPoints := (*[1 << 30]C.PointStruct)(unsafe.Pointer(points))[:int(count):int(count)]

	for i := 0; i < int(count); i++ {
		goPoints[i] = pointStructFromC(&cPoints[i])
	}

	if err := col.Upsert(goPoints); err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_GENERAL, err.Error())
		}
		return C.GOVECTOR_ERROR_GENERAL
	}

	return C.GOVECTOR_OK
}

func govector_collection_search(
	handle C.CollectionHandle,
	queryVector *C.float,
	vectorDim C.int,
	topK C.int,
	results **C.ScoredPoint,
	resultCount *C.int,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	col := collectionRegistry[uint64(uintptr(handle))]
	registryMu.RUnlock()

	if col == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_NOT_FOUND, "collection not found")
		}
		return C.GOVECTOR_ERROR_NOT_FOUND
	}

	// Convert query vector
	goVector := make([]float32, int(vectorDim))
	if queryVector != nil {
		cVector := (*[1 << 30]C.float)(unsafe.Pointer(queryVector))[:int(vectorDim):int(vectorDim)]
		for i := 0; i < int(vectorDim); i++ {
			goVector[i] = float32(cVector[i])
		}
	}

	// Execute search
	scoredPoints, err := col.Search(goVector, nil, int(topK))
	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_GENERAL, err.Error())
		}
		return C.GOVECTOR_ERROR_GENERAL
	}

	// Allocate C memory for results
	count := len(scoredPoints)
	if count > 0 && results != nil {
		cResults := (*C.ScoredPoint)(C.malloc(C.size_t(count) * C.sizeof_ScoredPoint))
		if cResults == nil {
			if errInfo != nil {
				*errInfo = newError(C.GOVECTOR_ERROR_MEMORY_ALLOC, "failed to allocate memory for results")
			}
			return C.GOVECTOR_ERROR_MEMORY_ALLOC
		}

		cResultsSlice := (*[1 << 30]C.ScoredPoint)(unsafe.Pointer(cResults))[:count:count]
		for i, point := range scoredPoints {
			cResultsSlice[i] = scoredPointToC(point)
		}

		*results = cResults
		*resultCount = C.int(count)
	} else {
		if results != nil {
			*results = nil
		}
		if resultCount != nil {
			*resultCount = 0
		}
	}

	return C.GOVECTOR_OK
}

func govector_collection_delete(
	handle C.CollectionHandle,
	ids **C.char,
	count C.int,
	deletedCount *C.int,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	col := collectionRegistry[uint64(uintptr(handle))]
	registryMu.RUnlock()

	if col == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_NOT_FOUND, "collection not found")
		}
		return C.GOVECTOR_ERROR_NOT_FOUND
	}

	// Convert IDs array
	goIDs := make([]string, int(count))
	if ids != nil {
		cIDs := (*[1 << 30]*C.char)(unsafe.Pointer(ids))[:int(count):int(count)]
		for i := 0; i < int(count); i++ {
			goIDs[i] = C.GoString(cIDs[i])
		}
	}

	// Execute delete
	n, err := col.Delete(goIDs, nil)
	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOVECTOR_ERROR_GENERAL, err.Error())
		}
		return C.GOVECTOR_ERROR_GENERAL
	}

	if deletedCount != nil {
		*deletedCount = C.int(n)
	}
	return C.GOVECTOR_OK
}

func govector_collection_count(handle C.CollectionHandle) C.int {
	registryMu.RLock()
	col := collectionRegistry[uint64(uintptr(handle))]
	registryMu.RUnlock()

	if col == nil {
		return 0
	}

	return C.int(col.Count())
}

// ============================================================================
// Memory Management API Implementation
// ============================================================================

func govector_points_free(points *C.PointStruct, count C.int) {
	if points == nil {
		return
	}

	cPoints := (*[1 << 30]C.PointStruct)(unsafe.Pointer(points))[:int(count):int(count)]
	for i := 0; i < int(count); i++ {
		if cPoints[i].id != nil {
			C.free(unsafe.Pointer(cPoints[i].id))
		}
		if cPoints[i].vector != nil {
			C.free(unsafe.Pointer(cPoints[i].vector))
		}
		if cPoints[i].payload_json != nil {
			C.free(unsafe.Pointer(cPoints[i].payload_json))
		}
	}
	C.free(unsafe.Pointer(points))
}

func govector_scored_points_free(results *C.ScoredPoint, count C.int) {
	if results == nil {
		return
	}

	cResults := (*[1 << 30]C.ScoredPoint)(unsafe.Pointer(results))[:int(count):int(count)]
	for i := 0; i < int(count); i++ {
		if cResults[i].id != nil {
			C.free(unsafe.Pointer(cResults[i].id))
		}
		if cResults[i].payload_json != nil {
			C.free(unsafe.Pointer(cResults[i].payload_json))
		}
	}
	C.free(unsafe.Pointer(results))
}

func govector_error_free(errInfo *C.ErrorInfo) {
	if errInfo != nil && errInfo.message != nil {
		C.free(unsafe.Pointer(errInfo.message))
	}
}
