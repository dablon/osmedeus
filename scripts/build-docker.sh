#!/bin/bash

# Docker build automation script for AI-enhanced Osmedeus
set -e

# Configuration
REGISTRY=${DOCKER_REGISTRY:-"osmedeus"}
VERSION=${VERSION:-"latest"}
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
PUSH_TO_REGISTRY=${PUSH_TO_REGISTRY:-"false"}
BUILD_PLATFORM=${BUILD_PLATFORM:-"linux/amd64"}
CACHE_FROM=${CACHE_FROM:-""}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Function to push image to registry
push_image() {
    local service=$1
    local tag=$2
    
    log "Pushing ${tag} to registry..."
    docker push "${tag}" || error "Failed to push ${tag}"
    log "Successfully pushed ${tag}"
}

# Function to check if Docker buildx is available
check_buildx() {
    if ! docker buildx version >/dev/null 2>&1; then
        warn "Docker buildx not available, falling back to regular docker build"
        return 1
    fi
    return 0
}

# Function to setup buildx builder if needed
setup_buildx() {
    if check_buildx; then
        # Create and use a new builder instance for multi-platform builds
        docker buildx create --name osmedeus-builder --use 2>/dev/null || true
        docker buildx inspect --bootstrap >/dev/null 2>&1 || warn "Failed to bootstrap buildx builder"
    fi
}

# Function to build a Docker image with enhanced features
build_image() {
    local service=$1
    local dockerfile=$2
    local context=$3
    local tag="${REGISTRY}/${service}:${VERSION}"
    local cache_args=""
    
    log "Building ${service} image..."
    
    # Add cache arguments if specified
    if [ -n "${CACHE_FROM}" ]; then
        cache_args="--cache-from ${CACHE_FROM}/${service}:latest"
    fi
    
    # Build with buildx for multi-platform support
    docker buildx build \
        --platform "${BUILD_PLATFORM}" \
        --file "${dockerfile}" \
        --tag "${tag}" \
        --build-arg BUILD_DATE="${BUILD_DATE}" \
        --build-arg GIT_COMMIT="${GIT_COMMIT}" \
        --build-arg VERSION="${VERSION}" \
        ${cache_args} \
        --load \
        "${context}" || error "Failed to build ${service}"
    
    log "Successfully built ${tag}"
    
    # Tag as latest if version is not latest
    if [ "${VERSION}" != "latest" ]; then
        docker tag "${tag}" "${REGISTRY}/${service}:latest"
        log "Tagged ${REGISTRY}/${service}:latest"
    fi
    
    # Push to registry if enabled
    if [ "${PUSH_TO_REGISTRY}" = "true" ]; then
        push_image "${service}" "${tag}"
        if [ "${VERSION}" != "latest" ]; then
            push_image "${service}" "${REGISTRY}/${service}:latest"
        fi
    fi
}

# Function to build GPU-enabled image
build_gpu_image() {
    local service=$1
    local dockerfile=$2
    local context=$3
    local tag="${REGISTRY}/${service}:${VERSION}-gpu"
    local cache_args=""
    
    log "Building ${service} GPU image..."
    
    # Add cache arguments if specified
    if [ -n "${CACHE_FROM}" ]; then
        cache_args="--cache-from ${CACHE_FROM}/${service}:latest-gpu"
    fi
    
    if check_buildx; then
        docker buildx build \
            --platform "${BUILD_PLATFORM}" \
            --file "${dockerfile}" \
            --target gpu-base \
            --tag "${tag}" \
            --build-arg BUILD_DATE="${BUILD_DATE}" \
            --build-arg GIT_COMMIT="${GIT_COMMIT}" \
            --build-arg VERSION="${VERSION}" \
            ${cache_args} \
            --load \
            "${context}" || error "Failed to build ${service} GPU image"
    else
        docker build \
            --file "${dockerfile}" \
            --target gpu-base \
            --tag "${tag}" \
            --build-arg BUILD_DATE="${BUILD_DATE}" \
            --build-arg GIT_COMMIT="${GIT_COMMIT}" \
            --build-arg VERSION="${VERSION}" \
            "${context}" || error "Failed to build ${service} GPU image"
    fi
    
    log "Successfully built ${tag}"
    
    # Push to registry if enabled
    if [ "${PUSH_TO_REGISTRY}" = "true" ]; then
        push_image "${service}" "${tag}"
    fi
}

# Function to build CPU-only image
build_cpu_image() {
    local service=$1
    local dockerfile=$2
    local context=$3
    local tag="${REGISTRY}/${service}:${VERSION}-cpu"
    local cache_args=""
    
    log "Building ${service} CPU image..."
    
    # Add cache arguments if specified
    if [ -n "${CACHE_FROM}" ]; then
        cache_args="--cache-from ${CACHE_FROM}/${service}:latest-cpu"
    fi
    
    if check_buildx; then
        docker buildx build \
            --platform "${BUILD_PLATFORM}" \
            --file "${dockerfile}" \
            --target cpu-base \
            --tag "${tag}" \
            --build-arg BUILD_DATE="${BUILD_DATE}" \
            --build-arg GIT_COMMIT="${GIT_COMMIT}" \
            --build-arg VERSION="${VERSION}" \
            ${cache_args} \
            --load \
            "${context}" || error "Failed to build ${service} CPU image"
    else
        docker build \
            --file "${dockerfile}" \
            --target cpu-base \
            --tag "${tag}" \
            --build-arg BUILD_DATE="${BUILD_DATE}" \
            --build-arg GIT_COMMIT="${GIT_COMMIT}" \
            --build-arg VERSION="${VERSION}" \
            "${context}" || error "Failed to build ${service} CPU image"
    fi
    
    log "Successfully built ${tag}"
    
    # Push to registry if enabled
    if [ "${PUSH_TO_REGISTRY}" = "true" ]; then
        push_image "${service}" "${tag}"
    fi
}

# Main build process
main() {
    log "Starting Docker build process for AI-enhanced Osmedeus"
    log "Registry: ${REGISTRY}"
    log "Version: ${VERSION}"
    log "Build Date: ${BUILD_DATE}"
    log "Git Commit: ${GIT_COMMIT}"
    log "Platform: ${BUILD_PLATFORM}"
    log "Push to Registry: ${PUSH_TO_REGISTRY}"
    
    # Setup buildx for multi-platform builds
    setup_buildx
    
    # Build main Osmedeus application
    build_image "osmedeus" "Dockerfile" "."
    
    # Build AI services
    build_image "ai-asset-classifier" "docker/ai-asset-classifier/Dockerfile" "docker/ai-asset-classifier"
    build_image "behavioral-analyzer" "docker/behavioral-analyzer/Dockerfile" "docker/behavioral-analyzer"
    build_image "exploit-generator" "docker/exploit-generator/Dockerfile" "docker/exploit-generator"
    
    # Build ML inference service with GPU and CPU variants
    build_gpu_image "ml-inference" "docker/ml-inference/Dockerfile" "docker/ml-inference"
    build_cpu_image "ml-inference" "docker/ml-inference/Dockerfile" "docker/ml-inference"
    
    log "All images built successfully!"
    
    # Display built images
    log "Built images:"
    docker images | grep "${REGISTRY}" | head -20
    
    # Cleanup buildx builder if we created it
    if check_buildx; then
        docker buildx rm osmedeus-builder 2>/dev/null || true
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --push)
            PUSH_TO_REGISTRY="true"
            shift
            ;;
        --platform)
            BUILD_PLATFORM="$2"
            shift 2
            ;;
        --cache-from)
            CACHE_FROM="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --registry REGISTRY    Docker registry prefix (default: osmedeus)"
            echo "  --version VERSION      Image version tag (default: latest)"
            echo "  --push                 Push images to registry after building"
            echo "  --platform PLATFORM   Target platform for build (default: linux/amd64)"
            echo "  --cache-from REGISTRY  Use registry for build cache"
            echo "  --help                 Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 --version v1.0.0 --push"
            echo "  $0 --registry myregistry.com/osmedeus --platform linux/amd64,linux/arm64"
            echo "  $0 --cache-from myregistry.com/osmedeus"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Run main function
main