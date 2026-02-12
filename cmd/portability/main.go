package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"docgen/config"
	"docgen/internal/portability"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config.InitLogging()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <export|import> [flags]\n", os.Args[0])
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "export":
		exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
		outFile := exportCmd.String("o", "export.json", "output file path")
		includeDeleted := exportCmd.Bool("include-deleted", false, "include soft-deleted records")
		exportCmd.Parse(os.Args[2:])
		runExport(*outFile, *includeDeleted)

	case "import":
		importCmd := flag.NewFlagSet("import", flag.ExitOnError)
		inFile := importCmd.String("i", "", "input file path (required)")
		dryRun := importCmd.Bool("dry-run", false, "validate without writing to database")
		importCmd.Parse(os.Args[2:])
		if *inFile == "" {
			fmt.Fprintf(os.Stderr, "Error: -i flag is required\n")
			os.Exit(1)
		}
		runImport(*inFile, *dryRun)

	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\nUsage: %s <export|import> [flags]\n", subcommand, os.Args[0])
		os.Exit(1)
	}
}

func connectDB(ctx context.Context) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, config.PostgreSQLConnString())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	return pool
}

func runExport(outFile string, includeDeleted bool) {
	ctx := context.Background()
	pool := connectDB(ctx)
	defer pool.Close()

	bundle, err := portability.Export(ctx, pool, includeDeleted)
	if err != nil {
		slog.Error("export failed", "error", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		slog.Error("failed to marshal JSON", "error", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outFile, data, 0644); err != nil {
		slog.Error("failed to write file", "error", err)
		os.Exit(1)
	}

	slog.Info("export complete", "file", outFile, "size_bytes", len(data))
}

func runImport(inFile string, dryRun bool) {
	ctx := context.Background()
	pool := connectDB(ctx)
	defer pool.Close()

	data, err := os.ReadFile(inFile)
	if err != nil {
		slog.Error("failed to read file", "error", err)
		os.Exit(1)
	}

	var bundle portability.ExportBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		slog.Error("failed to parse JSON", "error", err)
		os.Exit(1)
	}

	slog.Info("parsed bundle",
		"version", bundle.Version,
		"exported_at", bundle.ExportedAt,
		"section_rows", len(bundle.SectionRows),
		"sections", len(bundle.Sections),
		"pages", len(bundle.Pages),
		"images", len(bundle.Images),
	)

	if err := portability.Validate(&bundle); err != nil {
		slog.Error("bundle validation failed", "error", err)
		os.Exit(1)
	}
	slog.Info("bundle validation passed")

	if dryRun {
		slog.Info("dry run complete, no changes written")
		return
	}

	if err := portability.Import(ctx, pool, &bundle); err != nil {
		slog.Error("import failed", "error", err)
		os.Exit(1)
	}
}
