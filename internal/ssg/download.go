package ssg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var defaultHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// VendorOptions defines configuration for acquiring third-party frontend dependencies.
type VendorOptions struct {
	Offline   bool
	VendorDir string
	CacheDir  string
	WarnFunc  func(format string, args ...any)
}

// ResolveUserCacheDir returns the standard cross-platform path for caching vendor assets.
func ResolveUserCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "qst", "vendor"), nil
}

func hasKaTeX(dir string) bool {
	stat, err := os.Stat(filepath.Join(dir, "katex", "katex.min.css"))
	return err == nil && !stat.IsDir()
}

func hasMermaid(dir string) bool {
	stat, err := os.Stat(filepath.Join(dir, "mermaid", "mermaid.min.js"))
	return err == nil && !stat.IsDir()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("source directory %s does not exist", src)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		return copyFile(path, targetPath)
	})
}

// extractTar unpacks tar entries into destDir with Zip Slip protection.
func extractTar(tr *tar.Reader, destDir string) error {
	cleanDest := filepath.Clean(destDir)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(cleanDest, header.Name)
		cleanTarget := filepath.Clean(target)

		// Zip Slip / Path Traversal protection
		if cleanTarget != cleanDest && !strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		if header.FileInfo().IsDir() {
			if mkdirErr := os.MkdirAll(cleanTarget, 0o755); mkdirErr != nil {
				return mkdirErr
			}
			continue
		}

		if mkdirErr := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); mkdirErr != nil {
			return mkdirErr
		}
		f, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
	}
	return nil
}

// downloadKaTeX downloads and extracts the KaTeX tarball to destDir.
func downloadKaTeX(destDir string) error {
	if hasKaTeX(destDir) {
		return nil
	}

	resp, err := defaultHTTPClient.Get("https://github.com/KaTeX/KaTeX/releases/download/v0.16.9/katex.tar.gz")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download KaTeX: HTTP %d", resp.StatusCode)
	}

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)
	return extractTar(tr, destDir)
}

// downloadMermaid downloads the standalone mermaid bundle to destDir/mermaid/mermaid.min.js.
func downloadMermaid(destDir string) error {
	if hasMermaid(destDir) {
		return nil
	}
	target := filepath.Join(destDir, "mermaid", "mermaid.min.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	resp, err := defaultHTTPClient.Get("https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Mermaid: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, resp.Body)
	return err
}

// EnsureVendorAssets resolves and installs KaTeX and Mermaid dependencies into outDir/assets/vendor.
// In offline mode, it checks outDir, custom vendorDir, and user cache without network access.
// In online mode, it caches downloaded assets to user cache for future offline reuse.
func EnsureVendorAssets(outDir string, opts VendorOptions) error {
	vendorDir := filepath.Join(outDir, "assets", "vendor")
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		return err
	}

	hasKatexDest := hasKaTeX(vendorDir)
	hasMermaidDest := hasMermaid(vendorDir)
	if hasKatexDest && hasMermaidDest {
		return nil
	}

	// 1. Try copying from user-supplied vendorDir
	if opts.VendorDir != "" {
		if !hasKatexDest && hasKaTeX(opts.VendorDir) {
			if err := copyDir(filepath.Join(opts.VendorDir, "katex"), filepath.Join(vendorDir, "katex")); err == nil {
				hasKatexDest = true
			}
		}
		if !hasMermaidDest && hasMermaid(opts.VendorDir) {
			if err := copyDir(filepath.Join(opts.VendorDir, "mermaid"), filepath.Join(vendorDir, "mermaid")); err == nil {
				hasMermaidDest = true
			}
		}
	}

	// 2. Try copying from persistent user cache (~/.cache/qst/vendor)
	cacheDir := opts.CacheDir
	if cacheDir == "" {
		cacheDir, _ = ResolveUserCacheDir()
	}
	if cacheDir != "" {
		if !hasKatexDest && hasKaTeX(cacheDir) {
			if err := copyDir(filepath.Join(cacheDir, "katex"), filepath.Join(vendorDir, "katex")); err == nil {
				hasKatexDest = true
			}
		}
		if !hasMermaidDest && hasMermaid(cacheDir) {
			if err := copyDir(filepath.Join(cacheDir, "mermaid"), filepath.Join(vendorDir, "mermaid")); err == nil {
				hasMermaidDest = true
			}
		}
	}

	// If all assets are satisfied, we're done
	if hasKatexDest && hasMermaidDest {
		return nil
	}

	// 3. Handle Offline Mode
	if opts.Offline {
		if opts.WarnFunc != nil {
			opts.WarnFunc("Warning: offline mode enabled and vendor assets (KaTeX / Mermaid) were not found; math and diagrams will not render locally.\n")
		}
		return nil
	}

	// 4. Online Mode: Download missing assets to cache (if available) and target vendorDir
	destForDownload := vendorDir
	if cacheDir != "" {
		if err := os.MkdirAll(cacheDir, 0o755); err == nil {
			destForDownload = cacheDir
		}
	}

	if !hasKatexDest {
		if err := downloadKaTeX(destForDownload); err != nil {
			if destForDownload != vendorDir {
				if err2 := downloadKaTeX(vendorDir); err2 != nil {
					return fmt.Errorf("failed to download KaTeX: %w", err)
				}
			} else {
				return fmt.Errorf("failed to download KaTeX: %w", err)
			}
		} else if destForDownload != vendorDir {
			_ = copyDir(filepath.Join(destForDownload, "katex"), filepath.Join(vendorDir, "katex"))
		}
	}

	if !hasMermaidDest {
		if err := downloadMermaid(destForDownload); err != nil {
			if destForDownload != vendorDir {
				if err2 := downloadMermaid(vendorDir); err2 != nil {
					return fmt.Errorf("failed to download Mermaid: %w", err)
				}
			} else {
				return fmt.Errorf("failed to download Mermaid: %w", err)
			}
		} else if destForDownload != vendorDir {
			_ = copyDir(filepath.Join(destForDownload, "mermaid"), filepath.Join(vendorDir, "mermaid"))
		}
	}

	return nil
}

// DownloadAssets orchestrates the downloading of KaTeX and Mermaid to outDir/assets/vendor.
func DownloadAssets(outDir string) error {
	return EnsureVendorAssets(outDir, VendorOptions{})
}
