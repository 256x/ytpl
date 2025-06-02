#!/bin/bash
set -e  # Exit on error
set -x  # Print debug information

# Save current directory
PROJECT_ROOT=$(pwd)

# Get version information
VERSION=$(grep 'const Version' cmd/root.go | awk -F'"' '{print $2}')
OUTPUT_DIR="${PROJECT_ROOT}/release"

# Create output directory
mkdir -p ${OUTPUT_DIR}

echo "Building ytpl version: ${VERSION}"

# Define target platforms (Linux only since we use mpv player)
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
)

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    # Split platform into OS and architecture
    OS=$(echo ${platform} | cut -d'/' -f1)
    ARCH=$(echo ${platform} | cut -d'/' -f2)
    
    # Set output filename
    OUTPUT_NAME="ytpl-${VERSION}-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    echo "\n=== Building for ${OS}/${ARCH} ==="
    echo "Output: ${OUTPUT_DIR}/${OUTPUT_NAME}"
    
    # Execute build command
    cd "${PROJECT_ROOT}"  # Always return to project root
    set +e  # Temporarily disable error detection
    env CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -v -ldflags="-s -w" -o "${OUTPUT_DIR}/${OUTPUT_NAME}" .
    BUILD_STATUS=$?
    set -e  # Re-enable error detection
    
    # Check if build was successful
    if [ ${BUILD_STATUS} -ne 0 ]; then
        echo "!! Error building for ${OS}/${ARCH} (status: ${BUILD_STATUS}) !!"
        continue  # Continue with next platform even if build fails
    fi
    
    # Compression
    echo "Compressing..."
    cd "${OUTPUT_DIR}"  # Change to output directory
    
    if [ "$OS" = "windows" ]; then
        zip "${OUTPUT_NAME}.zip" "${OUTPUT_NAME}" && \
        rm -f "${OUTPUT_NAME}" && \
        echo "Created: ${OUTPUT_DIR}/${OUTPUT_NAME}.zip"
    else
        tar -czf "${OUTPUT_NAME}.tar.gz" "${OUTPUT_NAME}" && \
        rm -f "${OUTPUT_NAME}" && \
        echo "Created: ${OUTPUT_DIR}/${OUTPUT_NAME}.tar.gz"
    fi
    
    cd - > /dev/null  # Return to original directory
done

echo "\n=== Build Summary ==="
echo "Build completed. Output files in ${OUTPUT_DIR}:"
ls -lh ${OUTPUT_DIR}/ | grep -v "^total"

echo "\nTo create a release, run the following commands:"
echo "git tag -a v${VERSION} -m \"Release ${VERSION}\""
echo "git push origin v${VERSION}"
