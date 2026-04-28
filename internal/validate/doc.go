// Package validate implements pre-flight validation rules for secret diffs
// produced by the diff package.
//
// Validation is intentionally separate from diffing and patching so that
// rules can evolve independently and be composed into any command pipeline.
//
// Usage:
//
//	changes := diff.Compute(src, dst)
//	result  := validate.Validate(changes)
//	if !result.OK() {
//		validate.Fprint(os.Stderr, result, false)
//		os.Exit(1)
//	}
package validate
