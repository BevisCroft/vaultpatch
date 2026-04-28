// Command vaultpatch is a CLI tool to diff and apply HashiCorp Vault secret
// changes across environments safely.
//
// Usage:
//
//	vaultpatch diff        compare secrets between two paths or snapshots
//	vaultpatch apply       apply a diff/patch to a target Vault path
//	vaultpatch snapshot    capture a point-in-time snapshot of a Vault path
//	vaultpatch promote     promote secrets from one environment to another
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/youorg/vaultpatch/internal/audit"
	"github.com/youorg/vaultpatch/internal/config"
	"github.com/youorg/vaultpatch/internal/patch"
	"github.com/youorg/vaultpatch/internal/promote"
	"github.com/youorg/vaultpatch/internal/snapshot"
	"github.com/youorg/vaultpatch/internal/vault"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "vaultpatch",
		Short: "Diff and apply HashiCorp Vault secret changes safely",
		Long: `vaultpatch helps you manage Vault secret changes across environments.

It supports diffing paths, capturing snapshots, applying patches, and
promoting secrets between environments — all with dry-run and audit support.`,
		SilenceUsage: true,
	}

	root.AddCommand(
		newSnapshotCmd(),
		newPromoteCmd(),
		newApplyCmd(),
	)

	return root
}

// newSnapshotCmd returns the "snapshot" sub-command.
func newSnapshotCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "snapshot <path>",
		Short: "Capture a snapshot of secrets under a Vault path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.FromEnv()
			if err != nil {
				return err
			}
			client, err := vault.NewClient(cfg)
			if err != nil {
				return err
			}
			snap, err := snapshot.Capture(context.Background(), client, args[0])
			if err != nil {
				return fmt.Errorf("capture: %w", err)
			}
			if err := snapshot.Save(snap, outputFile); err != nil {
				return fmt.Errorf("save: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "snapshot saved to %s (%d secrets)\n", outputFile, len(snap.Secrets))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "snapshot.json", "file to write the snapshot to")
	return cmd
}

// newPromoteCmd returns the "promote" sub-command.
func newPromoteCmd() *cobra.Command {
	var dryRun, overwrite bool
	var auditLog string

	cmd := &cobra.Command{
		Use:   "promote <src-path> <dst-path>",
		Short: "Promote secrets from a source path to a destination path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.FromEnv()
			if err != nil {
				return err
			}
			cfg.DryRun = dryRun
			client, err := vault.NewClient(cfg)
			if err != nil {
				return err
			}
			auditor, err := audit.New(auditLog)
			if err != nil {
				return fmt.Errorf("audit: %w", err)
			}
			promoter := promote.New(client, cfg, auditor)
			return promoter.Run(context.Background(), args[0], args[1], overwrite)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing to Vault")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing keys in the destination")
	cmd.Flags().StringVar(&auditLog, "audit-log", "", "path to append audit entries (default: stderr)")
	return cmd
}

// newApplyCmd returns the "apply" sub-command.
func newApplyCmd() *cobra.Command {
	var dryRun bool
	var auditLog string

	cmd := &cobra.Command{
		Use:   "apply <snapshot-file> <dst-path>",
		Short: "Apply a snapshot file to a Vault path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.FromEnv()
			if err != nil {
				return err
			}
			cfg.DryRun = dryRun
			client, err := vault.NewClient(cfg)
			if err != nil {
				return err
			}
			auditor, err := audit.New(auditLog)
			if err != nil {
				return fmt.Errorf("audit: %w", err)
			}
			applier := patch.NewApplier(client, cfg, auditor)
			snap, err := snapshot.Load(args[0])
			if err != nil {
				return fmt.Errorf("load snapshot: %w", err)
			}
			report, err := applier.Apply(context.Background(), snap, args[1])
			if err != nil {
				return err
			}
			patch.FprintReport(cmd.OutOrStdout(), report)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing to Vault")
	cmd.Flags().StringVar(&auditLog, "audit-log", "", "path to append audit entries (default: stderr)")
	return cmd
}
