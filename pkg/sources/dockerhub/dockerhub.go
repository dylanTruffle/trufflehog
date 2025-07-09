package dockerhub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trufflesecurity/trufflehog/v3/pkg/common"
	"github.com/trufflesecurity/trufflehog/v3/pkg/context"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/sourcespb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources/docker"
)

const SourceType = sourcespb.SourceType_SOURCE_TYPE_DOCKERHUB

type Source struct {
	name        string
	sourceId    sources.SourceID
	jobId       sources.JobID
	verify      bool
	concurrency int
	conn        sourcespb.DockerHub
	httpClient  *http.Client
	
	sources.Progress
	sources.CommonSourceUnitUnmarshaller
}

// Ensure the Source satisfies the interfaces at compile time.
var _ sources.Source = (*Source)(nil)
var _ sources.SourceUnitUnmarshaller = (*Source)(nil)

// Type returns the type of source.
func (s *Source) Type() sourcespb.SourceType {
	return SourceType
}

func (s *Source) SourceID() sources.SourceID {
	return s.sourceId
}

func (s *Source) JobID() sources.JobID {
	return s.jobId
}

// DockerHub API response structures
type DockerHubRepository struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	PullCount   int64  `json:"pull_count"`
}

type DockerHubRepositoriesResponse struct {
	Count    int                   `json:"count"`
	Next     *string               `json:"next"`
	Previous *string               `json:"previous"`
	Results  []DockerHubRepository `json:"results"`
}

type DockerHubTag struct {
	Name        string    `json:"name"`
	FullSize    int64     `json:"full_size"`
	LastUpdated time.Time `json:"last_updated"`
	Digest      string    `json:"digest"`
}

type DockerHubTagsResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []DockerHubTag `json:"results"`
}

// Init initializes the source.
func (s *Source) Init(ctx context.Context, name string, jobId sources.JobID, sourceId sources.SourceID, verify bool, connection *anypb.Any, concurrency int) error {
	s.name = name
	s.sourceId = sourceId
	s.jobId = jobId
	s.verify = verify
	s.concurrency = concurrency
	s.httpClient = common.RetryableHTTPClientTimeout(60)

	if err := anypb.UnmarshalTo(connection, &s.conn, proto.UnmarshalOptions{}); err != nil {
		return fmt.Errorf("error unmarshalling connection: %w", err)
	}

	return nil
}

// Chunks emits data over a channel that is decoded and scanned for secrets.
func (s *Source) Chunks(ctx context.Context, chunksChan chan *sources.Chunk, _ ...sources.ChunkingTarget) error {
	ctx = context.WithValue(ctx, "source_type", s.Type())
	ctx = context.WithValue(ctx, "source_name", s.name)

	// Discover all repositories and tags
	images, err := s.enumerateImages(ctx)
	if err != nil {
		return fmt.Errorf("failed to enumerate images: %w", err)
	}

	ctx.Logger().Info("discovered Docker images", "count", len(images))

	// Scan each discovered image using the existing docker scanner
	workers := new(errgroup.Group)
	workers.SetLimit(s.concurrency)

	scanErrs := sources.NewScanErrors()
	for _, image := range images {
		image := image
		workers.Go(func() error {
			if common.IsDone(ctx) {
				return nil
			}

			if err := s.scanImage(ctx, image, chunksChan); err != nil {
				scanErrs.Add(fmt.Errorf("failed to scan image %s: %w", image, err))
			}
			return nil
		})
	}

	_ = workers.Wait()
	if scanErrs.Count() > 0 {
		ctx.Logger().V(2).Info("scan errors", "errors", scanErrs.String())
	}

	return nil
}

// enumerateImages discovers all repositories and their tags from DockerHub
func (s *Source) enumerateImages(ctx context.Context) ([]string, error) {
	var allImages []string
	var mu sync.Mutex

	workers := new(errgroup.Group)
	workers.SetLimit(s.concurrency)

	// Enumerate repositories for specified organizations
	for _, org := range s.conn.Organizations {
		org := org
		workers.Go(func() error {
			repos, err := s.getRepositoriesForOrganization(ctx, org)
			if err != nil {
				return fmt.Errorf("failed to get repositories for org %s: %w", org, err)
			}

			for _, repo := range repos {
				repoName := fmt.Sprintf("%s/%s", org, repo.Name)
				if s.shouldIncludeRepository(repoName) {
					images, err := s.getImagesForRepository(ctx, repoName)
					if err != nil {
						ctx.Logger().Error(err, "failed to get images for repository", "repo", repoName)
						continue
					}

					mu.Lock()
					allImages = append(allImages, images...)
					mu.Unlock()
				}
			}
			return nil
		})
	}

	// Add explicitly specified repositories
	for _, repo := range s.conn.Repositories {
		repo := repo
		workers.Go(func() error {
			if s.shouldIncludeRepository(repo) {
				images, err := s.getImagesForRepository(ctx, repo)
				if err != nil {
					return fmt.Errorf("failed to get images for repository %s: %w", repo, err)
				}

				mu.Lock()
				allImages = append(allImages, images...)
				mu.Unlock()
			}
			return nil
		})
	}

	if err := workers.Wait(); err != nil {
		return nil, err
	}

	return allImages, nil
}

// shouldIncludeRepository checks if a repository should be included based on include/exclude filters
func (s *Source) shouldIncludeRepository(repo string) bool {
	// Check exclude list first
	for _, excludePattern := range s.conn.IgnoreRepos {
		if strings.Contains(repo, excludePattern) {
			return false
		}
	}

	// If include list is specified, check if repo matches
	if len(s.conn.IncludeRepos) > 0 {
		for _, includePattern := range s.conn.IncludeRepos {
			if strings.Contains(repo, includePattern) {
				return true
			}
		}
		return false
	}

	return true
}

// getRepositoriesForOrganization fetches all repositories for a given organization
func (s *Source) getRepositoriesForOrganization(ctx context.Context, org string) ([]DockerHubRepository, error) {
	var allRepos []DockerHubRepository
	pageSize := 100
	page := 1

	for {
		url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/?page_size=%d&page=%d", 
			url.QueryEscape(org), pageSize, page)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if err := s.setAuthHeaders(req); err != nil {
			return nil, fmt.Errorf("failed to set auth headers: %w", err)
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var response DockerHubRepositoriesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allRepos = append(allRepos, response.Results...)

		// Check if there are more pages
		if response.Next == nil || len(response.Results) == 0 {
			break
		}
		page++
	}

	return allRepos, nil
}

// getImagesForRepository fetches all tags for a repository and returns fully qualified image names
func (s *Source) getImagesForRepository(ctx context.Context, repo string) ([]string, error) {
	var allImages []string
	pageSize := 100
	page := 1
	maxTags := int(s.conn.MaxTags)
	if maxTags == 0 {
		maxTags = 100 // default limit
	}

	tagCount := 0

	for {
		url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/?page_size=%d&page=%d", 
			url.QueryEscape(repo), pageSize, page)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if err := s.setAuthHeaders(req); err != nil {
			return nil, fmt.Errorf("failed to set auth headers: %w", err)
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var response DockerHubTagsResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for _, tag := range response.Results {
			if tagCount >= maxTags {
				return allImages, nil
			}
			
			imageName := fmt.Sprintf("%s:%s", repo, tag.Name)
			allImages = append(allImages, imageName)
			tagCount++
		}

		// Check if there are more pages and we haven't hit our limit
		if response.Next == nil || len(response.Results) == 0 || tagCount >= maxTags {
			break
		}
		page++
	}

	return allImages, nil
}

// setAuthHeaders sets appropriate authentication headers based on the credential type
func (s *Source) setAuthHeaders(req *http.Request) error {
	switch cred := s.conn.GetCredential().(type) {
	case *sourcespb.DockerHub_BasicAuth:
		req.SetBasicAuth(cred.BasicAuth.Username, cred.BasicAuth.Password)
	case *sourcespb.DockerHub_Token:
		req.Header.Set("Authorization", "Bearer "+cred.Token)
	case *sourcespb.DockerHub_Unauthenticated:
		// No authentication needed
	default:
		return fmt.Errorf("unknown credential type: %T", s.conn.Credential)
	}
	return nil
}

// scanImage uses the existing docker scanner to scan a specific image
func (s *Source) scanImage(ctx context.Context, imageName string, chunksChan chan *sources.Chunk) error {
	// Create a docker connection configuration for this specific image
	dockerConn := &sourcespb.Docker{
		Images: []string{imageName},
	}

	// Set authentication based on our DockerHub credentials
	switch cred := s.conn.GetCredential().(type) {
	case *sourcespb.DockerHub_BasicAuth:
		dockerConn.Credential = &sourcespb.Docker_BasicAuth{
			BasicAuth: cred.BasicAuth,
		}
	case *sourcespb.DockerHub_Token:
		dockerConn.Credential = &sourcespb.Docker_BearerToken{
			BearerToken: cred.Token,
		}
	case *sourcespb.DockerHub_Unauthenticated:
		dockerConn.Credential = &sourcespb.Docker_Unauthenticated{}
	default:
		return fmt.Errorf("unknown credential type: %T", s.conn.Credential)
	}

	// Marshall the docker connection
	var conn anypb.Any
	if err := anypb.MarshalFrom(&conn, dockerConn, proto.MarshalOptions{}); err != nil {
		return fmt.Errorf("failed to marshal docker connection: %w", err)
	}

	// Create and initialize a docker source
	dockerSource := &docker.Source{}
	sourceName := fmt.Sprintf("%s-docker-%s", s.name, imageName)
	if err := dockerSource.Init(ctx, sourceName, s.jobId, s.sourceId, s.verify, &conn, 1); err != nil {
		return fmt.Errorf("failed to initialize docker source: %w", err)
	}

	// Scan the image
	return dockerSource.Chunks(ctx, chunksChan)
}