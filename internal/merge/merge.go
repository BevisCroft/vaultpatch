package merge

import (
	"fmt"

	"github.com/your-org/vaultpatch/internal/vault"
)

// Strategy controls how conflicting keys are resolved.
type Strategy string

const (
	StrategyOurs   Strategy = "ours"   // keep destination value on conflict
	StrategyTheirs Strategy = "theirs" // overwrite with source value on conflict
	StrategyError  Strategy = "error"  // return error on conflict
)

// Result holds the outcome of a single path merge.
type Result struct {
	Path      string
	Merged    map[string]string
	Conflicts []string
	Err       error
	DryRun    bool
}

// Merger reads secrets from two paths and merges them.
type Merger struct {
	client   *vault.Client
	strategy Strategy
	dryRun   bool
}

// New creates a Merger with the given client, strategy, and dry-run flag.
func New(client *vault.Client, strategy Strategy, dryRun bool) *Merger {
	return &Merger{client: client, strategy: strategy, dryRun: dryRun}
}

// Apply merges secrets from srcPath into dstPath using the configured strategy.
func (m *Merger) Apply(srcPath, dstPath string) Result {
	res := Result{Path: dstPath, DryRun: m.dryRun}

	src, err := m.client.ReadSecret(srcPath)
	if err != nil {
		res.Err = fmt.Errorf("read source %s: %w", srcPath, err)
		return res
	}

	dst, err := m.client.ReadSecret(dstPath)
	if err != nil {
		res.Err = fmt.Errorf("read destination %s: %w", dstPath, err)
		return res
	}

	merged := make(map[string]string, len(dst))
	for k, v := range dst {
		merged[k] = v
	}

	for k, sv := range src {
		if dv, exists := merged[k]; exists && dv != sv {
			res.Conflicts = append(res.Conflicts, k)
			switch m.strategy {
			case StrategyError:
				res.Err = fmt.Errorf("conflict on key %q: use --strategy=ours or --strategy=theirs", k)
				return res
			case StrategyTheirs:
				merged[k] = sv
			// ours: keep existing dst value — no-op
			}
		} else {
			merged[k] = sv
		}
	}

	res.Merged = merged

	if !m.dryRun {
		if err := m.client.WriteSecret(dstPath, merged); err != nil {
			res.Err = fmt.Errorf("write %s: %w", dstPath, err)
			return res
		}
	}

	return res
}
