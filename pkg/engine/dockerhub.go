package engine

import (
	"runtime"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trufflesecurity/trufflehog/v3/pkg/context"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/credentialspb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/sourcespb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources/dockerhub"
)

// DockerHubConfig defines the configuration for a DockerHub source.
type DockerHubConfig struct {
	Organizations []string
	Repositories  []string
	IgnoreRepos   []string
	IncludeRepos  []string
	MaxTags       int32
	Username      string
	Password      string
	Token         string
}

// ScanDockerHub scans a given DockerHub connection.
func (e *Engine) ScanDockerHub(ctx context.Context, c DockerHubConfig) error {
	connection := &sourcespb.DockerHub{
		Organizations: c.Organizations,
		Repositories:  c.Repositories,
		IgnoreRepos:   c.IgnoreRepos,
		IncludeRepos:  c.IncludeRepos,
		MaxTags:       c.MaxTags,
	}

	// Set authentication based on provided credentials
	switch {
	case len(c.Token) > 0:
		connection.Credential = &sourcespb.DockerHub_Token{Token: c.Token}
	case len(c.Username) > 0 && len(c.Password) > 0:
		connection.Credential = &sourcespb.DockerHub_BasicAuth{
			BasicAuth: &credentialspb.BasicAuth{
				Username: c.Username,
				Password: c.Password,
			},
		}
	default:
		connection.Credential = &sourcespb.DockerHub_Unauthenticated{
			Unauthenticated: &credentialspb.Unauthenticated{},
		}
	}

	var conn anypb.Any
	err := anypb.MarshalFrom(&conn, connection, proto.MarshalOptions{})
	if err != nil {
		ctx.Logger().Error(err, "failed to marshal dockerhub connection")
		return err
	}

	sourceName := "trufflehog - dockerhub"
	sourceID, jobID, _ := e.sourceManager.GetIDs(ctx, sourceName, dockerhub.SourceType)

	dockerhubSource := &dockerhub.Source{}
	if err := dockerhubSource.Init(ctx, sourceName, jobID, sourceID, true, &conn, runtime.NumCPU()); err != nil {
		return err
	}
	_, err = e.sourceManager.Run(ctx, sourceName, dockerhubSource)
	return err
}