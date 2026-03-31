#!/bin/bash
# GoVector CAPI Build Script
# 
# This script builds CGO/SWIG bindings for multiple languages.
# You don't need to understand C++ or SWIG - just run this script!
#
# Usage:
#   ./build.sh              # Build everything (default)
#   ./build.sh python       # Build Python bindings only
#   ./build.sh java         # Build Java bindings only
#   ./build.sh csharp       # Build C# bindings only
#   ./build.sh clean        # Clean build files
#   ./build.sh help         # Show help

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
check_dependencies() {
    print_info "Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    
    if ! command -v swig &> /dev/null; then
        print_error "SWIG is not installed."
        print_info "Install with:"
        print_info "  macOS: brew install swig"
        print_info "  Ubuntu: sudo apt-get install swig"
        exit 1
    fi
    
    if ! command -v gcc &> /dev/null; then
        print_error "GCC is not installed."
        exit 1
    fi
    
    if [[ "$1" == "python" ]] && ! command -v python3 &> /dev/null; then
        print_error "Python 3 is not installed."
        exit 1
    fi
    
    print_success "All dependencies are available"
}

# Build Go static library
build_go_lib() {
    print_info "Building Go static library..."
    
    cd "$(dirname "$0")/.."
    go build -buildmode=c-archive -o capi/libgovector_c.a ./capi
    
    print_success "Go static library built: capi/libgovector_c.a"
}

# Generate SWIG wrapper
generate_swig_wrapper() {
    local lang=$1
    print_info "Generating SWIG wrapper for ${lang}..."
    
    cd "$(dirname "$0")"
    
    case $lang in
        python)
            swig -cgo -python -py3 -o govector_wrap.cxx govector.i
            print_success "Python wrapper generated: govector_wrap.cxx"
            ;;
        java)
            swig -cgo -java -o govector_wrap.cxx govector.i
            print_success "Java wrapper generated: govector_wrap.cxx"
            ;;
        csharp)
            swig -cgo -csharp -o govector_wrap.cxx govector.i
            print_success "C# wrapper generated: govector_wrap.cxx"
            ;;
        *)
            print_error "Unsupported language: ${lang}"
            exit 1
            ;;
    esac
}

# Build Python module
build_python_module() {
    print_info "Compiling Python module..."
    
    cd "$(dirname "$0")"
    
    # Get Python include paths
    PYTHON_INCLUDE=$(python3-config --includes 2>/dev/null || echo "-I/usr/include/python3.8")
    PYTHON_LIBS=$(python3-config --ldflags 2>/dev/null || echo "")
    
    # Compile shared library
    g++ -fPIC -shared govector_wrap.cxx -o _govector.so \
        ${PYTHON_INCLUDE} \
        -L. -lgovector_c \
        ${PYTHON_LIBS}
    
    print_success "Python module compiled: _govector.so"
    print_info "You can now use: import govector"
}

# Build Java module
build_java_module() {
    print_info "Compiling Java module..."
    
    cd "$(dirname "$0")"
    
    # Compile shared library
    g++ -fPIC -shared govector_wrap.cxx -o libgovector_java.dylib \
        -L. -lgovector_c
    
    print_success "Java native library compiled: libgovector_java.dylib"
    print_info "Use the generated Java files in your project"
}

# Build C# module
build_csharp_module() {
    print_info "Compiling C# module..."
    
    cd "$(dirname "$0")"
    
    # Compile shared library
    g++ -fPIC -shared govector_wrap.cxx -o libgovector_csharp.dylib \
        -L. -lgovector_c
    
    print_success "C# native library compiled: libgovector_csharp.dylib"
    print_info "Use the generated C# files in your project"
}

# Run tests
run_tests() {
    print_info "Running tests..."
    
    cd "$(dirname "$0")"
    
    if command -v python3 &> /dev/null; then
        if [ -f "_govector.so" ] || [ -f "govector.py" ]; then
            python3 -c "import govector; print('Python binding test passed!')" && \
            print_success "Tests passed" || \
            print_error "Tests failed"
        else
            print_warning "Python module not found. Run './build.sh python' first."
        fi
    fi
}

# Clean build files
clean() {
    print_info "Cleaning build files..."
    
    cd "$(dirname "$0")"
    
    rm -f govector_wrap.cxx govector.py _govector.so *.o *.a *.dylib *.so
    rm -rf __pycache__ *.pyc
    rm -f *.class *.jar
    
    print_success "Clean complete"
}

# Show help
show_help() {
    echo "GoVector CAPI Build Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  all          - Build everything (default)"
    echo "  python       - Build Python bindings"
    echo "  java         - Build Java bindings"
    echo "  csharp       - Build C# bindings"
    echo "  test         - Run tests"
    echo "  clean        - Clean build files"
    echo "  help         - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 all           # Build everything"
    echo "  $0 python        # Build Python bindings only"
    echo "  $0 clean && $0 all  # Clean and rebuild"
    echo ""
    echo "Output Files:"
    echo "  libgovector_c.a     - Go static library"
    echo "  govector_wrap.cxx   - SWIG generated wrapper"
    echo "  _govector.so        - Python shared library"
    echo "  govector.py         - Python wrapper module"
    echo ""
}

# Main function
main() {
    local command=${1:-all}
    
    case $command in
        all)
            check_dependencies
            build_go_lib
            generate_swig_wrapper python
            build_python_module
            run_tests
            print_success "Build complete! You can now use GoVector from Python."
            ;;
        python)
            check_dependencies python
            build_go_lib
            generate_swig_wrapper python
            build_python_module
            print_success "Python bindings built successfully!"
            ;;
        java)
            check_dependencies java
            build_go_lib
            generate_swig_wrapper java
            build_java_module
            print_success "Java bindings built successfully!"
            ;;
        csharp)
            check_dependencies csharp
            build_go_lib
            generate_swig_wrapper csharp
            build_csharp_module
            print_success "C# bindings built successfully!"
            ;;
        test)
            run_tests
            ;;
        clean)
            clean
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: ${command}"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
