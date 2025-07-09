package main

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trufflesecurity/trufflehog/v3/pkg/context"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/credentialspb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/sourcespb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources/dockerhub"
)

func main() {
	// Initialize a context
	ctx := context.Background()
	
	// Create a DockerHub connection for testing
	conn := &sourcespb.DockerHub{
		Repositories: []string{"library/hello-world"},
		MaxTags:      1,
		Credential:   &sourcespb.DockerHub_Unauthenticated{
			Unauthenticated: &credentialspb.Unauthenticated{},
		},
	}
	
	// Marshal the connection
	var connection anypb.Any
	if err := anypb.MarshalFrom(&connection, conn, proto.MarshalOptions{}); err != nil {
		fmt.Printf("Failed to marshal connection: %v\n", err)
		os.Exit(1)
	}
	
	// Create and initialize the DockerHub source
	source := &dockerhub.Source{}
	if err := source.Init(ctx, "test-dockerhub", 1, 1, true, &connection, 1); err != nil {
		fmt.Printf("Failed to initialize DockerHub source: %v\n", err)
		os.Exit(1)
	}
	
	// Test source methods
	fmt.Printf("Source Type: %d\n", source.Type())
	fmt.Printf("Source ID: %d\n", source.SourceID())
	fmt.Printf("Job ID: %d\n", source.JobID())
	
	// Create a channel for chunks
	chunksChan := make(chan *sources.Chunk, 10)
	
	// Test enumeration (this would normally scan the images)
	fmt.Println("Testing DockerHub source enumeration...")
	
	// Run chunks in a goroutine
	go func() {
		defer close(chunksChan)
		if err := source.Chunks(ctx, chunksChan); err != nil {
			fmt.Printf("Error during chunks: %v\n", err)
		}
	}()
	
	// Count the chunks
	chunkCount := 0
	for chunk := range chunksChan {
		chunkCount++
		fmt.Printf("Received chunk from %s (size: %d bytes)\n", chunk.SourceName, len(chunk.Data))
		// Only show first few chunks to avoid spam
		if chunkCount >= 5 {
			break
		}
	}
	
	fmt.Printf("DockerHub source test completed successfully! Processed %d chunks.\n", chunkCount)
}