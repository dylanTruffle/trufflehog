package docker

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trufflesecurity/trufflehog/v3/pkg/common"
	"github.com/trufflesecurity/trufflehog/v3/pkg/context"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/source_metadatapb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/sourcespb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources"
)

const SourceType = sourcespb.SourceType_SOURCE_TYPE_DOCKER

type Source struct {
	name        string
	sourceId    sources.SourceID
	jobId       sources.JobID
	verify      bool
	concurrency int
	conn        sourcespb.Docker
	layerCache  *LayerCache
	sources.Progress
	sources.CommonSourceUnitUnmarshaller
}

// Ensure the Source satisfies the interfaces at compile time.
var _ sources.Source = (*Source)(nil)
var _ sources.SourceUnitUnmarshaller = (*Source)(nil)

// Type returns the type of source.
// It is used for matching source types in configuration and job input.
func (s *Source) Type() sourcespb.SourceType {
	return SourceType
}

func (s *Source) SourceID() sources.SourceID {
	return s.sourceId
}

func (s *Source) JobID() sources.JobID {
	return s.jobId
}

// Init initializes the source.
func (s *Source) Init(_ context.Context, name string, jobId sources.JobID, sourceId sources.SourceID, verify bool, connection *anypb.Any, concurrency int) error {
	s.name = name
	s.sourceId = sourceId
	s.jobId = jobId
	s.verify = verify
	s.concurrency = concurrency

	// Initialize layer cache with a reasonable size
	s.layerCache = NewLayerCache(1000)

	// Reset metrics for this source at initialization time.
	dockerImagesScanned.WithLabelValues(s.name).Set(0)
	dockerLayersScanned.WithLabelValues(s.name).Set(0)

	if err := anypb.UnmarshalTo(connection, &s.conn, proto.UnmarshalOptions{}); err != nil {
		return fmt.Errorf("error unmarshalling connection: %w", err)
	}

	return nil
}

// SetLayerCache sets the layer cache for this source (used for sharing cache across sources)
func (s *Source) SetLayerCache(cache *LayerCache) {
	s.layerCache = cache
}

type imageInfo struct {
	image v1.Image
	base  string
	tag   string
}

type historyEntryInfo struct {
	index       int
	entry       v1.History
	layerDigest string
	base        string
	tag         string
}

type layerInfo struct {
	digest v1.Hash
	base   string
	tag    string
}

// Chunks emits data over a channel that is decoded and scanned for secrets.
func (s *Source) Chunks(ctx context.Context, chunksChan chan *sources.Chunk, _ ...sources.ChunkingTarget) error {
	ctx = context.WithValues(ctx, "source_type", s.Type(), "source_name", s.name)

	workers := new(errgroup.Group)
	workers.SetLimit(s.concurrency)

	scanErrs := sources.NewScanErrors()
	for _, image := range s.conn.GetImages() {
		image := image
		workers.Go(func() error {
			if common.IsDone(ctx) {
				return nil
			}

			imgInfo, err := s.processImage(ctx, image)
			if err != nil {
				scanErrs.Add(err)
				return nil
			}

			ctx = context.WithValues(ctx, "image", imgInfo.base, "tag", imgInfo.tag)

			ctx.Logger().V(2).Info("scanning image history")

			historyEntries, err := getHistoryEntries(ctx, imgInfo)
			if err != nil {
				scanErrs.Add(err)
				return nil
			}

			for _, historyEntry := range historyEntries {
				if err := s.processHistoryEntry(ctx, historyEntry, chunksChan); err != nil {
					scanErrs.Add(err)
					return nil
				}
				dockerHistoryEntriesScanned.WithLabelValues(s.name).Inc()
			}

			ctx.Logger().V(2).Info("scanning image layers")

			layers, err := imgInfo.image.Layers()
			if err != nil {
				scanErrs.Add(err)
				return nil
			}

			for _, layer := range layers {
				if err := s.processLayer(ctx, layer, imgInfo, chunksChan); err != nil {
					scanErrs.Add(err)
					return nil
				}
				dockerLayersScanned.WithLabelValues(s.name).Inc()
			}

			dockerImagesScanned.WithLabelValues(s.name).Inc()

			return nil
		})
	}
	_ = workers.Wait()
	if scanErrs.Count() > 0 {
		ctx.Logger().V(2).Info("scan errors", "errors", scanErrs.String())
	}

	// Log cache statistics
	if s.layerCache != nil {
		hits, misses, size := s.layerCache.GetStats()
		ctx.Logger().Info("layer cache statistics", 
			"hits", hits, 
			"misses", misses, 
			"size", size, 
			"hit_rate", float64(hits)/float64(hits+misses)*100.0)
	}

	return nil
}

// processImage processes an individual image and prepares it for further processing.
func (s *Source) processImage(ctx context.Context, image string) (imageInfo, error) {
	var (
		imgInfo   imageInfo
		hasDigest bool
		imageName name.Reference
	)

	remoteOpts, err := s.remoteOpts()
	if err != nil {
		return imgInfo, err
	}

	const filePrefix = "file://"
	if strings.HasPrefix(image, filePrefix) {
		image = strings.TrimPrefix(image, filePrefix)
		imgInfo.base = image
		imgInfo.image, err = tarball.ImageFromPath(image, nil)
		if err != nil {
			return imgInfo, err
		}
	} else {
		imgInfo.base, imgInfo.tag, hasDigest = baseAndTagFromImage(image)

		if hasDigest {
			imageName, err = name.NewDigest(image)
		} else {
			imageName, err = name.NewTag(image)
		}
		if err != nil {
			return imgInfo, err
		}

		imgInfo.image, err = remote.Image(imageName, remoteOpts...)
		if err != nil {
			return imgInfo, err
		}
	}

	ctx.Logger().WithValues("image", imgInfo.base, "tag", imgInfo.tag).V(2).Info("scanning image")

	return imgInfo, nil
}

// getHistoryEntries collates an image's configuration history together with the
// corresponding layer digests for any non-empty layers.
func getHistoryEntries(ctx context.Context, imgInfo imageInfo) ([]historyEntryInfo, error) {
	config, err := imgInfo.image.ConfigFile()
	if err != nil {
		return nil, err
	}

	layers, err := imgInfo.image.Layers()
	if err != nil {
		return nil, err
	}

	history := config.History
	entries := make([]historyEntryInfo, len(history))

	layerIndex := 0
	for historyIndex, entry := range history {
		e := historyEntryInfo{
			base:  imgInfo.base,
			tag:   imgInfo.tag,
			entry: entry,
			index: historyIndex,
		}

		// Associate with a layer if possible -- failing to do this will not affect
		// the scan, just remove some traceability.
		if !entry.EmptyLayer {
			if layerIndex < len(layers) {
				digest, err := layers[layerIndex].Digest()

				if err == nil {
					e.layerDigest = digest.String()
				} else {
					ctx.Logger().V(2).Error(err, "cannot associate layer with history entry: layer digest failed",
						"layerIndex", layerIndex, "historyIndex", historyIndex)
				}
			} else {
				ctx.Logger().V(2).Info("cannot associate layer with history entry: no correlated layer exists at this index",
					"layerIndex", layerIndex, "historyIndex", historyIndex)
			}

			layerIndex++
		}

		entries[historyIndex] = e
	}

	return entries, nil
}

// processHistoryEntry processes a history entry from the image configuration metadata.
func (s *Source) processHistoryEntry(ctx context.Context, historyInfo historyEntryInfo, chunksChan chan *sources.Chunk) error {
	// Make up an identifier for this entry that is moderately sensible. There is
	// no file name to use here, so the path tries to be a little descriptive.
	entryPath := fmt.Sprintf("image-metadata:history:%d:created-by", historyInfo.index)

	chunk := &sources.Chunk{
		SourceType: s.Type(),
		SourceName: s.name,
		SourceID:   s.SourceID(),
		SourceMetadata: &source_metadatapb.MetaData{
			Data: &source_metadatapb.MetaData_Docker{
				Docker: &source_metadatapb.Docker{
					File:  entryPath,
					Image: historyInfo.base,
					Tag:   historyInfo.tag,
					Layer: historyInfo.layerDigest,
				},
			},
		},
		Verify: s.verify,
		Data:   []byte(historyInfo.entry.CreatedBy),
	}

	ctx.Logger().V(2).Info("scanning image history entry", "index", historyInfo.index, "layer", historyInfo.layerDigest)

	return common.CancellableWrite(ctx, chunksChan, chunk)
}

// processLayer processes an individual layer of an image.
func (s *Source) processLayer(ctx context.Context, layer v1.Layer, imgInfo imageInfo, chunksChan chan *sources.Chunk) error {
	layerInfo := layerInfo{
		base: imgInfo.base,
		tag:  imgInfo.tag,
	}

	var err error
	layerInfo.digest, err = layer.Digest()
	if err != nil {
		return err
	}

	ctx.Logger().WithValues("layer", layerInfo.digest.String()).V(2).Info("scanning layer")

	// Check if this layer is already cached
	if cachedResult, found := s.layerCache.Get(layerInfo.digest); found {
		ctx.Logger().WithValues("layer", layerInfo.digest.String()).V(2).Info("layer found in cache, using cached results")
		
		// Send cached chunks to the output channel
		for _, chunk := range cachedResult.Chunks {
			// Create a copy of the chunk with updated metadata for this specific image
			chunkCopy := &sources.Chunk{
				SourceType:     chunk.SourceType,
				SourceName:     chunk.SourceName,
				SourceID:       chunk.SourceID,
				SourceMetadata: chunk.SourceMetadata,
				Verify:         chunk.Verify,
				Data:           chunk.Data,
			}
			
			// Update the image and tag in the metadata
			if dockerMeta := chunkCopy.SourceMetadata.GetDocker(); dockerMeta != nil {
				dockerMeta.Image = imgInfo.base
				dockerMeta.Tag = imgInfo.tag
				// Keep the same layer digest and file path
			}
			
			if err := common.CancellableWrite(ctx, chunksChan, chunkCopy); err != nil {
				return err
			}
		}
		
		return cachedResult.Error
	}

	ctx.Logger().WithValues("layer", layerInfo.digest.String()).V(2).Info("layer not in cache, scanning")

	// Layer not cached, scan it and store the results
	var chunks []*sources.Chunk
	var scanError error

	rc, err := layer.Compressed()
	if err != nil {
		scanError = err
		s.layerCache.Set(layerInfo.digest, &LayerScanResult{
			Chunks: chunks,
			Error:  scanError,
		})
		return err
	}
	defer rc.Close()

	gzipReader, err := gzip.NewReader(rc)
	if err != nil {
		scanError = err
		s.layerCache.Set(layerInfo.digest, &LayerScanResult{
			Chunks: chunks,
			Error:  scanError,
		})
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			scanError = err
			break
		}

		info := chunkProcessingInfo{size: header.Size, name: header.Name, reader: tarReader, layer: layerInfo}
		
		// Process the chunk and collect results
		fileChunks, err := s.processChunkForCache(ctx, info)
		if err != nil {
			scanError = err
			break
		}
		
		// Add chunks to our collection and send to output
		for _, chunk := range fileChunks {
			chunks = append(chunks, chunk)
			if err := common.CancellableWrite(ctx, chunksChan, chunk); err != nil {
				scanError = err
				break
			}
		}

		if scanError != nil {
			break
		}
	}

	// Cache the results
	s.layerCache.Set(layerInfo.digest, &LayerScanResult{
		Chunks: chunks,
		Error:  scanError,
	})

	return scanError
}

type chunkProcessingInfo struct {
	size   int64
	name   string
	reader io.Reader
	layer  layerInfo
}

// processChunk processes an individual chunk of a layer.
func (s *Source) processChunk(ctx context.Context, info chunkProcessingInfo, chunksChan chan *sources.Chunk) error {
	const filesizeLimitBytes int64 = 50 * 1024 * 1024 // 50MB
	if info.size > filesizeLimitBytes {
		ctx.Logger().V(4).Info("skipping large file", "file", info.name, "size", info.size)
		return nil
	}

	chunkReader := sources.NewChunkReader()
	chunkResChan := chunkReader(ctx, info.reader)

	for data := range chunkResChan {
		if err := data.Error(); err != nil {
			ctx.Logger().Error(err, "error reading chunk.")
			continue
		}

		chunk := &sources.Chunk{
			SourceType: s.Type(),
			SourceName: s.name,
			SourceID:   s.SourceID(),
			SourceMetadata: &source_metadatapb.MetaData{
				Data: &source_metadatapb.MetaData_Docker{
					Docker: &source_metadatapb.Docker{
						File:  "/" + info.name,
						Image: info.layer.base,
						Tag:   info.layer.tag,
						Layer: info.layer.digest.String(),
					},
				},
			},
			Verify: s.verify,
		}
		chunk.Data = data.Bytes()

		if err := common.CancellableWrite(ctx, chunksChan, chunk); err != nil {
			return err
		}
	}

	return nil
}

// processChunkForCache processes a chunk and returns all chunks as a slice for caching
func (s *Source) processChunkForCache(ctx context.Context, info chunkProcessingInfo) ([]*sources.Chunk, error) {
	const filesizeLimitBytes int64 = 50 * 1024 * 1024 // 50MB
	if info.size > filesizeLimitBytes {
		ctx.Logger().V(4).Info("skipping large file", "file", info.name, "size", info.size)
		return nil, nil
	}

	var chunks []*sources.Chunk
	chunkReader := sources.NewChunkReader()
	chunkResChan := chunkReader(ctx, info.reader)

	for data := range chunkResChan {
		if err := data.Error(); err != nil {
			ctx.Logger().Error(err, "error reading chunk.")
			continue
		}

		chunk := &sources.Chunk{
			SourceType: s.Type(),
			SourceName: s.name,
			SourceID:   s.SourceID(),
			SourceMetadata: &source_metadatapb.MetaData{
				Data: &source_metadatapb.MetaData_Docker{
					Docker: &source_metadatapb.Docker{
						File:  "/" + info.name,
						Image: info.layer.base,
						Tag:   info.layer.tag,
						Layer: info.layer.digest.String(),
					},
				},
			},
			Verify: s.verify,
		}
		chunk.Data = data.Bytes()
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (s *Source) remoteOpts() ([]remote.Option, error) {
	switch s.conn.GetCredential().(type) {
	case *sourcespb.Docker_Unauthenticated:
		return nil, nil
	case *sourcespb.Docker_BasicAuth:
		return []remote.Option{
			remote.WithAuth(&authn.Basic{
				Username: s.conn.GetBasicAuth().GetUsername(),
				Password: s.conn.GetBasicAuth().GetPassword(),
			}),
		}, nil
	case *sourcespb.Docker_BearerToken:
		return []remote.Option{
			remote.WithAuth(&authn.Bearer{
				Token: s.conn.GetBearerToken(),
			}),
		}, nil
	case *sourcespb.Docker_DockerKeychain:
		return []remote.Option{
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
		}, nil
	default:
		return nil, fmt.Errorf("unknown credential type: %T", s.conn.Credential)
	}
}

func baseAndTagFromImage(image string) (base, tag string, hasDigest bool) {
	if base, tag, hasDigest = extractDigest(image); hasDigest {
		return base, tag, true
	}

	base, tag = extractTagOrUseDefault(image)
	return base, tag, false
}

// extractDigest tries to split the image string on the digest delimiter.
// If successful, it means the image has a digest.
func extractDigest(image string) (base, tag string, hasDigest bool) {
	const digestDelim = "@"

	if parts := strings.SplitN(image, digestDelim, 2); len(parts) > 1 {
		return parts[0], parts[1], true
	}
	return "", "", false
}

// extractTagOrUseDefault extracts the tag from the image string.
// If no tag is found, it defaults to "latest".
func extractTagOrUseDefault(image string) (base, tag string) {
	const (
		tagDelim     = ":"
		regRepoDelim = "/"
	)

	parts := strings.Split(image, tagDelim)

	// Check if the last part is not a hostname with a port (for weak validation)
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelim) {
		return strings.Join(parts[:len(parts)-1], tagDelim), parts[len(parts)-1]
	}
	return image, "latest"
}
