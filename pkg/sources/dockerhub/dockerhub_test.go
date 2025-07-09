package dockerhub

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trufflesecurity/trufflehog/v3/pkg/context"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/sourcespb"
)

func TestSource_Init(t *testing.T) {
	ctx := context.Background()
	
	// Test unauthenticated initialization
	conn := &sourcespb.DockerHub{
		Repositories: []string{"library/hello-world"},
		MaxTags:      1,
		Credential:   &sourcespb.DockerHub_Unauthenticated{},
	}
	
	var connection anypb.Any
	err := anypb.MarshalFrom(&connection, conn, proto.MarshalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	
	source := &Source{}
	err = source.Init(ctx, "test-dockerhub", 1, 1, true, &connection, 1)
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify source type
	if source.Type() != SourceType {
		t.Errorf("expected source type %d, got %d", SourceType, source.Type())
	}
	
	// Verify source ID
	if source.SourceID() != 1 {
		t.Errorf("expected source ID 1, got %d", source.SourceID())
	}
	
	// Verify job ID
	if source.JobID() != 1 {
		t.Errorf("expected job ID 1, got %d", source.JobID())
	}
}

func TestSource_Chunks_DryRun(t *testing.T) {
	ctx := context.Background()
	
	// Test with a public repository that should exist
	conn := &sourcespb.DockerHub{
		Repositories: []string{"library/hello-world"},
		MaxTags:      1,
		Credential:   &sourcespb.DockerHub_Unauthenticated{},
	}
	
	var connection anypb.Any
	err := anypb.MarshalFrom(&connection, conn, proto.MarshalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	
	source := &Source{}
	err = source.Init(ctx, "test-dockerhub", 1, 1, true, &connection, 1)
	if err != nil {
		t.Fatal(err)
	}
	
	// Test that we can enumerate at least one image
	images, err := source.enumerateImages(ctx)
	if err != nil {
		t.Skip("Skipping test - unable to enumerate images (network issue or rate limit)")
	}
	
	if len(images) == 0 {
		t.Error("expected at least one image to be enumerated")
	}
}

func TestShouldIncludeRepository(t *testing.T) {
	tests := []struct {
		name         string
		repo         string
		ignoreRepos  []string
		includeRepos []string
		expected     bool
	}{
		{
			name:         "no filters",
			repo:         "user/repo",
			ignoreRepos:  []string{},
			includeRepos: []string{},
			expected:     true,
		},
		{
			name:         "ignored repo",
			repo:         "user/ignored",
			ignoreRepos:  []string{"ignored"},
			includeRepos: []string{},
			expected:     false,
		},
		{
			name:         "included repo",
			repo:         "user/included",
			ignoreRepos:  []string{},
			includeRepos: []string{"included"},
			expected:     true,
		},
		{
			name:         "not in include list",
			repo:         "user/notincluded",
			ignoreRepos:  []string{},
			includeRepos: []string{"included"},
			expected:     false,
		},
		{
			name:         "ignored takes precedence",
			repo:         "user/conflicted",
			ignoreRepos:  []string{"conflicted"},
			includeRepos: []string{"conflicted"},
			expected:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &Source{
				conn: sourcespb.DockerHub{
					IgnoreRepos:  tt.ignoreRepos,
					IncludeRepos: tt.includeRepos,
				},
			}
			
			result := source.shouldIncludeRepository(tt.repo)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}