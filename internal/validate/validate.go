// Package validate provides pre-flight checks for Vault secret diffs
// before they are applied to an environment.
package validate

import (
	"fmt"
	"strings"

	"github.com/your-org/vaultpatch/internal/diff"
)

// Issue represents a single validation problem found in a diff.
type Issue struct {
	Severity string // "error" or "warning"
	Path     string
	Message  string
}

func (i Issue) String() string {
	return fmt.Sprintf("[%s] %s: %s", strings.ToUpper(i.Severity), i.Path, i.Message)
}

// Result holds all issues found during validation.
type Result struct {
	Issues []Issue
}

// OK returns true when no error-level issues exist.
func (r *Result) OK() bool {
	for _, iss := range r.Issues {
		if iss.Severity == "error" {
			return false
		}
	}
	return true
}

// Validate inspects a slice of diff.Changes and returns a Result.
// Rules:
//   - A key whose value is an empty string is a warning.
//   - A key name containing whitespace is an error.
//   - Removing a key whose name ends with "_required" is an error.
func Validate(changes []diff.Change) *Result {
	res := &Result{}
	for _, c := range changes {
		if strings.ContainsAny(c.Key, " \t") {
			res.Issues = append(res.Issues, Issue{
				Severity: "error",
				Path:     c.Path + "/" + c.Key,
				Message:  "key name contains whitespace",
			})
		}
		if c.Op == diff.OpAdd || c.Op == diff.OpUpdate {
			if fmt.Sprintf("%v", c.NewValue) == "" {
				res.Issues = append(res.Issues, Issue{
					Severity: "warning",
					Path:     c.Path + "/" + c.Key,
					Message:  "new value is empty",
				})
			}
		}
		if c.Op == diff.OpRemove && strings.HasSuffix(c.Key, "_required") {
			res.Issues = append(res.Issues, Issue{
				Severity: "error",
				Path:     c.Path + "/" + c.Key,
				Message:  "removal of a _required key is not allowed",
			})
		}
	}
	return res
}
