package dockerhub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
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
	sharedCache *docker.LayerCache
	
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
	
	// Initialize shared cache for all images in this dockerhub scan
	s.sharedCache = docker.NewLayerCache(5000)
	// Temporarily disable cache for baseline comparison
	// s.sharedCache.SetEnabled(false)

	if err := anypb.UnmarshalTo(connection, &s.conn, proto.UnmarshalOptions{}); err != nil {
		return fmt.Errorf("error unmarshalling connection: %w", err)
	}

	// Validate configuration
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return nil
}

// validateConfig validates the DockerHub configuration for common issues
func (s *Source) validateConfig() error {
	// Check that we have at least one organization or repository
	if len(s.conn.Organizations) == 0 && len(s.conn.Repositories) == 0 {
		return fmt.Errorf("at least one organization or repository must be specified")
	}

	// Validate max tags is reasonable
	if s.conn.MaxTags < 0 {
		return fmt.Errorf("max_tags cannot be negative")
	}
	if s.conn.MaxTags > 1000 {
		// Note: We don't have access to ctx here, so we can't log. This is just a validation check.
		// Large max_tags values could cause memory issues but will be handled during execution.
	}

	// Validate repository names
	for _, repo := range s.conn.Repositories {
		if err := s.validateRepositoryName(repo); err != nil {
			return fmt.Errorf("invalid repository name %q: %w", repo, err)
		}
	}

	// Validate organization names
	for _, org := range s.conn.Organizations {
		if err := s.validateOrganizationName(org); err != nil {
			return fmt.Errorf("invalid organization name %q: %w", org, err)
		}
	}

	return nil
}

// validateRepositoryName validates a repository name format
func (s *Source) validateRepositoryName(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	if len(repo) > 255 {
		return fmt.Errorf("repository name too long (max 255 characters)")
	}
	// Basic validation for Docker repository name format
	if strings.Contains(repo, "..") || strings.HasPrefix(repo, "/") || strings.HasSuffix(repo, "/") {
		return fmt.Errorf("invalid repository name format")
	}
	return nil
}

// validateOrganizationName validates an organization name format
func (s *Source) validateOrganizationName(org string) error {
	if org == "" {
		return fmt.Errorf("organization name cannot be empty")
	}
	if len(org) > 255 {
		return fmt.Errorf("organization name too long (max 255 characters)")
	}
	// Basic validation for Docker organization name format
	if strings.Contains(org, "/") || strings.Contains(org, "..") {
		return fmt.Errorf("invalid organization name format")
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

	// Early return if no images found
	if len(images) == 0 {
		ctx.Logger().V(1).Info("no images found to scan")
		return nil
	}

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

	// Log overall cache statistics for the dockerhub scan
	if s.sharedCache != nil {
		hits, misses, size := s.sharedCache.GetStats()
		totalRequests := hits + misses
		hitRate := float64(0)
		if totalRequests > 0 {
			hitRate = float64(hits) / float64(totalRequests) * 100.0
		}
		ctx.Logger().Info("DockerHub scan cache statistics", 
			"total_images", len(images),
			"cache_hits", hits, 
			"cache_misses", misses, 
			"cache_size", size, 
			"hit_rate_percent", hitRate)
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
			if common.IsDone(ctx) {
				return nil
			}

			repos, err := s.getRepositoriesForOrganization(ctx, org)
			if err != nil {
				return fmt.Errorf("failed to get repositories for org %s: %w", org, err)
			}

			for _, repo := range repos {
				if common.IsDone(ctx) {
					return nil
				}

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
			if common.IsDone(ctx) {
				return nil
			}

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
// Now uses proper glob pattern matching instead of simple string contains
func (s *Source) shouldIncludeRepository(repo string) bool {
	// Check exclude list first using glob patterns
	for _, excludePattern := range s.conn.IgnoreRepos {
		if matched, _ := filepath.Match(excludePattern, repo); matched {
			return false
		}
		// Fallback to simple string contains for backwards compatibility
		if strings.Contains(repo, excludePattern) {
			return false
		}
	}

	// If include list is specified, check if repo matches using glob patterns
	if len(s.conn.IncludeRepos) > 0 {
		for _, includePattern := range s.conn.IncludeRepos {
			if matched, _ := filepath.Match(includePattern, repo); matched {
				return true
			}
			// Fallback to simple string contains for backwards compatibility
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
	maxPages := 100 // Prevent infinite pagination loops

	for page <= maxPages {
		if common.IsDone(ctx) {
			break
		}

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
		
		// Ensure response body is always closed
		func() {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				// Organization doesn't exist or has no repositories
				ctx.Logger().V(1).Info("organization not found or has no repositories", "org", org)
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				// Rate limited - could implement backoff here
				ctx.Logger().V(1).Info("rate limited by DockerHub API", "org", org)
				return
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // Limit error response size
				ctx.Logger().Error(fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body)), "failed to get repositories", "org", org)
				return
			}

			// Limit response size to prevent memory issues
			limitedReader := io.LimitReader(resp.Body, 10*1024*1024) // 10MB limit
			var response DockerHubRepositoriesResponse
			if err := json.NewDecoder(limitedReader).Decode(&response); err != nil {
				ctx.Logger().Error(err, "failed to decode response", "org", org)
				return
			}

			allRepos = append(allRepos, response.Results...)

			// Check if there are more pages
			if response.Next == nil || len(response.Results) == 0 {
				return
			}
		}()

		page++
	}

	if page > maxPages {
		ctx.Logger().V(1).Info("hit maximum page limit for organization", "org", org, "max_pages", maxPages)
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
	maxPages := 50 // Prevent infinite pagination loops

	tagCount := 0

	for page <= maxPages && tagCount < maxTags {
		if common.IsDone(ctx) {
			break
		}

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

		// Ensure response body is always closed
		shouldBreak := false
		func() {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				ctx.Logger().V(1).Info("repository not found or has no tags", "repo", repo)
				shouldBreak = true
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				ctx.Logger().V(1).Info("rate limited by DockerHub API", "repo", repo)
				shouldBreak = true
				return
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // Limit error response size
				ctx.Logger().Error(fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body)), "failed to get tags", "repo", repo)
				shouldBreak = true
				return
			}

			// Limit response size to prevent memory issues
			limitedReader := io.LimitReader(resp.Body, 10*1024*1024) // 10MB limit
			var response DockerHubTagsResponse
			if err := json.NewDecoder(limitedReader).Decode(&response); err != nil {
				ctx.Logger().Error(err, "failed to decode response", "repo", repo)
				shouldBreak = true
				return
			}

			for _, tag := range response.Results {
				if tagCount >= maxTags {
					shouldBreak = true
					return
				}
				
				// Validate tag name
				if tag.Name == "" || len(tag.Name) > 128 {
					ctx.Logger().V(2).Info("skipping invalid tag", "repo", repo, "tag", tag.Name)
					continue
				}
				
				imageName := fmt.Sprintf("%s:%s", repo, tag.Name)
				allImages = append(allImages, imageName)
				tagCount++
			}

			// Check if there are more pages and we haven't hit our limit
			if response.Next == nil || len(response.Results) == 0 || tagCount >= maxTags {
				shouldBreak = true
				return
			}
		}()

		if shouldBreak {
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
		if cred.BasicAuth == nil {
			return fmt.Errorf("basic auth credentials are nil")
		}
		req.SetBasicAuth(cred.BasicAuth.Username, cred.BasicAuth.Password)
	case *sourcespb.DockerHub_Token:
		if cred.Token == "" {
			return fmt.Errorf("token is empty")
		}
		req.Header.Set("Authorization", "Bearer "+cred.Token)
	case *sourcespb.DockerHub_Unauthenticated:
		// No authentication needed
	case nil:
		// Default to unauthenticated
	default:
		return fmt.Errorf("unknown credential type: %T", s.conn.Credential)
	}
	return nil
}

// scanImage uses the existing docker scanner to scan a specific image
func (s *Source) scanImage(ctx context.Context, imageName string, chunksChan chan *sources.Chunk) error {
	// Validate image name
	if imageName == "" {
		return fmt.Errorf("image name cannot be empty")
	}

	// Create a docker connection configuration for this specific image
	dockerConn := &sourcespb.Docker{
		Images: []string{imageName},
	}

	// Set authentication based on our DockerHub credentials
	switch cred := s.conn.GetCredential().(type) {
	case *sourcespb.DockerHub_BasicAuth:
		if cred.BasicAuth != nil {
			dockerConn.Credential = &sourcespb.Docker_BasicAuth{
				BasicAuth: cred.BasicAuth,
			}
		}
	case *sourcespb.DockerHub_Token:
		if cred.Token != "" {
			dockerConn.Credential = &sourcespb.Docker_BearerToken{
				BearerToken: cred.Token,
			}
		}
	case *sourcespb.DockerHub_Unauthenticated:
		dockerConn.Credential = &sourcespb.Docker_Unauthenticated{}
	default:
		// Default to unauthenticated for unknown types
		dockerConn.Credential = &sourcespb.Docker_Unauthenticated{}
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

	// Set the shared cache for this scan
	dockerSource.SetLayerCache(s.sharedCache)

	// Scan the image
	return dockerSource.Chunks(ctx, chunksChan)
}