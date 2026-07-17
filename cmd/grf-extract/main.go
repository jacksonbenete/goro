package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kivutar/goro/res"
)

func main() {
	output := flag.String("o", "", "output directory")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [-o output-dir] archive.grf\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	archivePath := flag.Arg(0)
	outDir := *output
	if outDir == "" {
		base := strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath))
		outDir = base
	}

	if err := extractGRF(archivePath, outDir); err != nil {
		fmt.Fprintf(os.Stderr, "grf-extract: %v\n", err)
		os.Exit(1)
	}
}

func extractGRF(archivePath, outDir string) error {
	archive, err := res.OpenGRF(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	root, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}

	var count int
	var totalBytes int64
	for _, name := range archive.Names() {
		target, err := extractPath(root, name)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		data, err := archive.ReadFile(name)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
		count++
		totalBytes += int64(len(data))
	}

	fmt.Printf("extracted %d files (%d bytes) to %s\n", count, totalBytes, root)
	return nil
}

func extractPath(root, grfName string) (string, error) {
	name := filepath.Clean(filepath.FromSlash(grfName))
	if name == "." || filepath.IsAbs(name) || strings.HasPrefix(name, ".."+string(filepath.Separator)) || name == ".." {
		return "", fmt.Errorf("unsafe path")
	}
	target := filepath.Join(root, name)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", fmt.Errorf("unsafe path")
	}
	return target, nil
}
