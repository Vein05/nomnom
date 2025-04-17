#!/bin/bash

# Set color codes for better visibility
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to build for macOS
build_mac() {
    local arch=$1
    echo "🔧 Starting macOS $arch build process..."

    # Clean any existing builds
    rm -f nomnom "nomnom-darwin-$arch.zip"

    # Build the macOS binary
    echo "🔨 Building macOS $arch binary..."
    GOOS=darwin GOARCH=$arch go build -v -o nomnom

    # Check if build was successful
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ macOS $arch build failed!${NC}"
        return 1
    fi

    # Create zip archive
    echo "📦 Creating zip archive..."
    if zip -q "nomnom-darwin-$arch.zip" nomnom LICENSE README.md config.example.json; then
        echo -e "${GREEN}✅ Successfully created: nomnom-darwin-$arch.zip${NC}"
        echo "📊 Archive size: $(du -h nomnom-darwin-$arch.zip | cut -f1)"
    else
        echo -e "${RED}❌ Failed to create zip archive${NC}"
        return 1
    fi

    # Clean up
    rm -f nomnom
    echo -e "${GREEN}✅ macOS $arch build completed successfully!${NC}"
}

# Build for Apple Silicon (M1/M2)
build_mac "arm64"

# Build for Intel Mac
build_mac "amd64"

# Function to build for Linux
build_linux() {
    local arch=$1
    echo "🔧 Starting Linux $arch build process..."

    # Clean any existing builds
    rm -f nomnom "nomnom-linux-$arch.zip"

    # Build the Linux binary
    echo "🔨 Building Linux $arch binary..."
    GOOS=linux GOARCH=$arch go build -v -o nomnom

    # Check if build was successful
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Linux $arch build failed!${NC}"
        return 1
    fi

    # Create zip archive
    echo "📦 Creating zip archive..."
    if zip -q "nomnom-linux-$arch.zip" nomnom LICENSE README.md config.example.json; then
        echo -e "${GREEN}✅ Successfully created: nomnom-linux-$arch.zip${NC}"
        echo "📊 Archive size: $(du -h nomnom-linux-$arch.zip | cut -f1)"
    else
        echo -e "${RED}❌ Failed to create zip archive${NC}"
        return 1
    fi

    # Clean up
    rm -f nomnom
    echo -e "${GREEN}✅ Linux $arch build completed successfully!${NC}"
}

# Build for Linux AMD64
build_linux "amd64"

# Build for Linux ARM64
build_linux "arm64"

echo "🔧 Starting Windows AMD64 build process..."

# Check if cross-compiler is installed
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo -e "${RED}Error: Windows cross-compiler not found${NC}"
    echo "To install on macOS, run: brew install mingw-w64"
    exit 1
fi

# Clean any existing builds
rm -f nomnom.exe nomnom-windows-amd64.zip

# Build the Windows binary
echo "🔨 Building Windows AMD64 binary..."
CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC=x86_64-w64-mingw32-gcc \
go build -v -o nomnom.exe

# Check if build was successful
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi

# Create zip archive
echo "📦 Creating zip archive..."
if zip -q "nomnom-windows-amd64.zip" nomnom.exe LICENSE README.md config.example.json; then
    echo -e "${GREEN}✅ Successfully created: nomnom-windows-amd64.zip${NC}"
else
    echo -e "${RED}❌ Failed to create zip archive${NC}"
    exit 1
fi

# Show final file size
echo "📊 Archive size: $(du -h nomnom-windows-amd64.zip | cut -f1)"

# Clean up
echo "🧹 Cleaning up..."
rm -f nomnom.exe
rm -f nomnom
echo -e "${GREEN}✅ Build process completed successfully!${NC}"
