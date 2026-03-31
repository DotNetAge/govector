# GoVector CAPI Makefile
# 
# This Makefile provides convenient build targets for CGO/SWIG bindings.
# You don't need to understand Make - just use the commands below!
#
# Quick Start:
#   make              - Build Python bindings (default)
#   make python       - Build Python bindings
#   make java         - Build Java bindings
#   make csharp       - Build C# bindings
#   make clean        - Clean all build files
#   make help         - Show all available targets
#
# For detailed instructions, see: ../docs/CGO_QUICKSTART.md

# Variables
SWIG = swig
GO = go
CC = gcc
CXX = g++

# Default target language
TARGET_LANG ?= python

# Input files
SWIG_INTERFACE = govector.i
C_HEADER = govector_c.h
GO_SOURCE = govector_c.go

# Output files
SWIG_WRAP_CXX = govector_wrap.cxx
PYTHON_MODULE = _govector.so
PYTHON_WRAPPER = govector.py

# Flags
SWIG_FLAGS = -cgo -$(TARGET_LANG)

# Phony targets
.PHONY: all clean python java csharp build test help

# Default target
all: $(TARGET_LANG)

# Python bindings
python: SWIG_FLAGS += -python -py3
python: $(SWIG_WRAP_CXX)
	@echo "Building Python module..."
	$(CXX) -fPIC -shared $(SWIG_WRAP_CXX) -o $(PYTHON_MODULE) \
		-I$$(python3-config --includes) \
		-L. -lgovector_c \
		$$(python3-config --ldflags)
	@echo "Python module built: $(PYTHON_MODULE)"

# Java bindings
java: SWIG_FLAGS += -java
java: $(SWIG_WRAP_CXX)
	@echo "Building Java native library..."
	$(CXX) -fPIC -shared $(SWIG_WRAP_CXX) -o libgovector_java.dylib \
		-L. -lgovector_c
	@echo "Java library built: libgovector_java.dylib"

# C# bindings
csharp: SWIG_FLAGS += -csharp
csharp: $(SWIG_WRAP_CXX)
	@echo "Building C# native library..."
	$(CXX) -fPIC -shared $(SWIG_WRAP_CXX) -o libgovector_csharp.dylib \
		-L. -lgovector_c
	@echo "C# library built: libgovector_csharp.dylib"

# Generate SWIG wrapper file
$(SWIG_WRAP_CXX): $(SWIG_INTERFACE) $(C_HEADER)
	@echo "Generating SWIG wrapper..."
	$(SWIG) $(SWIG_FLAGS) -o $@ $<

# Build Go static library
build:
	@echo "Building Go static library..."
	cd .. && $(GO) build -buildmode=c-archive -o capi/libgovector_c.a ./capi
	@echo "Go library built: capi/libgovector_c.a"

# Run tests
test:
	@echo "Running tests..."
	@if command -v python3 &> /dev/null && [ -f "$(PYTHON_MODULE)" ]; then \
		python3 -c "import govector; print('✓ Python binding works!')"; \
	else \
		echo "⚠ Python module not found. Run 'make python' first."; \
	fi

# Clean build files
clean:
	@echo "Cleaning build files..."
	rm -f $(SWIG_WRAP_CXX) $(PYTHON_WRAPPER) $(PYTHON_MODULE) *.o *.so *.a *.dylib
	rm -rf __pycache__ *.pyc *.class *.jar
	@echo "Clean complete"

# Show help
help:
	@echo "GoVector CAPI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              - Build Python bindings (default)"
	@echo "  make python       - Build Python bindings"
	@echo "  make java         - Build Java bindings"
	@echo "  make csharp       - Build C# bindings"
	@echo "  make build        - Build Go static library"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build files"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make clean && make python    # Clean rebuild of Python bindings"
	@echo "  make java                    # Build Java bindings"
	@echo ""
	@echo "For more information, see:"
	@echo "  ../docs/CGO_QUICKSTART.md"
	@echo "  ../docs/SWIG_INTEGRATION_GUIDE.md"
