// Package compare provides environment-to-environment secret comparison
// for vaultpatch, producing a structured report of keys that differ,
// are missing, or are present only in one environment.
package compare

import (
	"context"
	"fmt"
	"sort"

	"github.com/your-org/vaultpatch/internal/vault"
)

// Result holds the comparison outcome for a single secret path.
type Result struct {
	Path    string
	OnlyIn  map[string][]string // env name -> keys exclusive to that env
	Differ  []string            // keys present in both but with different values
}

// Comparer reads secrets from two named Vault environments and compares them.
type Comparer struct {
	srcClient *vault.Client
	dstClient *vault.Client
	srcName   string
	dstName   string
}

// New creates a Comparer for the given source and destination clients.
func New(srcName string, src *vault.Client, dstName string, dst *vault.Client) *Comparer {
	return &Comparer{
		srcClient: src,
		dstClient: dst,
		srcName:   srcName,
		dstName:   dstName,
	}
}

// Compare reads secrets at path from both environments and returns a Result.
func (c *Comparer) Compare(ctx context.Context, path string) (Result, error) {
	srcData, err := c.srcClient.ReadSecret(ctx, path)
	if err != nil {
		return Result{}, fmt.Errorf("read %s from %s: %w", path, c.srcName, err)
	}
	dstData, err := c.dstClient.ReadSecret(ctx, path)
	if err != nil {
		return Result{}, fmt.Errorf("read %s from %s: %w", path, c.dstName, err)
	}

	res := Result{
		Path:   path,
		OnlyIn: map[string][]string{c.srcName: {}, c.dstName: {}},
	}

	for k, sv := range srcData {
		dv, ok := dstData[k]
		if !ok {
			res.OnlyIn[c.srcName] = append(res.OnlyIn[c.srcName], k)
			continue
		}
		if fmt.Sprintf("%v", sv) != fmt.Sprintf("%v", dv) {
			res.Differ = append(res.Differ, k)
		}
	}
	for k := range dstData {
		if _, ok := srcData[k]; !ok {
			res.OnlyIn[c.dstName] = append(res.OnlyIn[c.dstName], k)
		}
	}

	sort.Strings(res.Differ)
	sort.Strings(res.OnlyIn[c.srcName])
	sort.Strings(res.OnlyIn[c.dstName])
	return res, nil
}
