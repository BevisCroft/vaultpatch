// Package archive implements soft-deletion of Vault secrets by relocating
// them to a timestamped archive path rather than permanently removing them.
//
// # Overview
//
// An [Archiver] copies secret data from the original path to a destination
// under a configurable archive root, then deletes the source. The archive
// path includes a UTC timestamp so multiple archive runs never collide:
//
//	<archiveRoot>/<20060102T150405Z>/<original-path>
//
// # Dry-run mode
//
// When constructed with dryRun=true the Archiver only reports what would
// happen without reading, writing, or deleting any secrets.
package archive
