#!/bin/bash

# Benchmark script for Docker layer scanning performance
# Usage: ./benchmark_docker_scan.sh <repository> <max_tags>

REPO=${1:-"library/hello-world"}
MAX_TAGS=${2:-5}

echo "Benchmarking Docker scan performance"
echo "Repository: $REPO"
echo "Max tags: $MAX_TAGS"
echo "================================"

# Build the application
echo "Building application..."
go build -o trufflehog-benchmark .

echo "Running benchmark..."
time ./trufflehog-benchmark dockerhub --repo "$REPO" --max-tags "$MAX_TAGS" --no-update --no-verification > /tmp/benchmark_results.txt 2>&1

echo "Results:"
echo "------"
cat /tmp/benchmark_results.txt | grep -E "(discovered Docker images|finished scanning|layers|images)"

echo ""
echo "Performance metrics:"
tail -1 /tmp/benchmark_results.txt

# Clean up
rm -f ./trufflehog-benchmark