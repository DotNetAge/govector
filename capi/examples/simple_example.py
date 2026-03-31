#!/usr/bin/env python3
"""
GoVector Python Binding - Simple Example

This example demonstrates basic usage of GoVector from Python.
You don't need to understand C++ or SWIG - just run this script!

Prerequisites:
  1. Build the Python bindings first:
     cd govector/capi
     ./build.sh python
  
  2. Make sure you're in the capi directory when running this script

Usage:
  python3 examples/simple_example.py
"""

import sys
import os

# Add current directory to path so we can import govector
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

try:
    import govector
except ImportError:
    print("❌ Error: govector module not found!")
    print("\nPlease build the Python bindings first:")
    print("  cd govector/capi")
    print("  ./build.sh python")
    sys.exit(1)


def main():
    """Main example function"""
    
    print("=" * 70)
    print("  GoVector Python Binding - Simple Example")
    print("=" * 70)
    
    # Initialize error structure
    error = govector.ErrorInfo()
    
    # Step 1: Create storage engine
    print("\n[Step 1] Creating storage engine...")
    storage = govector.govector_storage_new(b"example.db".decode(), error)
    
    if not storage:
        print(f"❌ Failed to create storage: {error.message.decode()}")
        govector.govector_error_free(error)
        return
    
    print("✓ Storage created successfully")
    
    # Step 2: Create HNSW parameters
    print("\n[Step 2] Configuring HNSW parameters...")
    hnsw_params = govector.HNSWParams()
    hnsw_params.m = 16
    hnsw_params.ef_construction = 200
    hnsw_params.ef_search = 50
    hnsw_params.k = 2
    print(f"✓ HNSW params: M={hnsw_params.m}, EF_CONSTRUCTION={hnsw_params.ef_construction}")
    
    # Step 3: Create collection
    print("\n[Step 3] Creating vector collection...")
    collection = govector.govector_collection_create(
        b"my_collection".decode(),
        3,  # Vector dimension (3D for this example)
        govector.DISTANCE_COSINE,  # Cosine similarity
        True,  # Use HNSW for fast search
        hnsw_params,
        storage,
        error
    )
    
    if not collection:
        print(f"❌ Failed to create collection: {error.message.decode()}")
        govector.govector_error_free(error)
        govector.govector_storage_free(storage)
        return
    
    print("✓ Collection created successfully")
    
    # Step 4: Get collection count
    print("\n[Step 4] Getting collection statistics...")
    count = govector.govector_collection_count(collection)
    print(f"✓ Collection contains {count} points")
    
    # Step 5: Clean up resources
    print("\n[Step 5] Cleaning up resources...")
    govector.govector_collection_free(collection)
    govector.govector_storage_close(storage)
    govector.govector_storage_free(storage)
    govector.govector_error_free(error)
    
    print("✓ All resources freed")
    
    print("\n" + "=" * 70)
    print("  Example completed successfully!")
    print("=" * 70)
    print("\nNext steps:")
    print("  - Try the advanced example: examples/demo.py")
    print("  - Read the documentation: ../docs/CGO_QUICKSTART.md")
    print("  - Check more examples: ../docs/SWIG_EXAMPLES_AND_BEST_PRACTICES.md")
    print()


if __name__ == "__main__":
    main()
