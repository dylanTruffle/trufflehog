package engine

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/trufflesecurity/trufflehog/v3/pkg/context"
    "github.com/trufflesecurity/trufflehog/v3/pkg/detectors"
    "github.com/trufflesecurity/trufflehog/v3/pkg/sources"
)

// discardPrinter is a no-op Printer implementation used in tests.
// It satisfies the engine.Printer interface but does nothing.
// This avoids cluttering test output while still exercising the dispatch path.
type discardPrinter struct{}

func (p *discardPrinter) Print(context.Context, *detectors.ResultWithMetadata) error { return nil }

// TestDockerHubConfigValidation ensures that calling ScanDockerHub with an
// empty configuration (no organisations or repositories) fails fast and does
// not attempt any network calls.
func TestDockerHubConfigValidation(t *testing.T) {
    ctx := context.Background()

    const defaultOutputBufferSize = 64
    sourceMgr := sources.NewManager(
        sources.WithBufferedOutput(defaultOutputBufferSize),
    )

    engCfg := Config{
        Concurrency:   1,
        Verify:        false,
        SourceManager: sourceMgr,
        Dispatcher:    NewPrinterDispatcher(new(discardPrinter)),
    }

    e, err := NewEngine(ctx, &engCfg)
    assert.NoError(t, err)

    e.Start(ctx)

    // Provide an empty DockerHubConfig – this should be rejected.
    dhCfg := DockerHubConfig{}
    err = e.ScanDockerHub(ctx, dhCfg)
    assert.Error(t, err, "expected error for missing orgs/repos")
}