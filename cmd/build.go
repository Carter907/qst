package cmd

import (
	"os"

	"github.com/Carter907/qst/internal/ssg"
	"github.com/spf13/cobra"
)

var (
	buildOut       string
	buildOffline   bool
	buildVendorDir string
)

var buildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Compile the knowledge graph into a static HTML website",
	Long: `Compile the knowledge graph into a static HTML website.

By default, the compiler fetches KaTeX and Mermaid vendor assets over HTTP and caches them in your local user cache directory (~/.cache/qst/vendor).
Use --offline to skip all network requests and compile using local assets.
Use --vendor-dir or the QST_VENDOR_DIR environment variable to provide custom pre-downloaded vendor assets.`,
	Example: `# Standard build with online assets
qst build . --out ./public

# Offline build (reuses cached assets or builds without external downloads)
qst build . --out ./public --offline

# Offline build using a specific vendor assets directory
qst build . --out ./public --offline --vendor-dir /path/to/vendor`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		vendorDir := buildVendorDir
		if vendorDir == "" {
			vendorDir = os.Getenv("QST_VENDOR_DIR")
		}

		cmd.Printf("Building static site from '%s' to '%s'...\n", dir, buildOut)
		err := ssg.BuildSiteWithOptions(ssg.BuildOptions{
			DirPath:   dir,
			OutDir:    buildOut,
			Offline:   buildOffline,
			VendorDir: vendorDir,
			WarnFunc: func(format string, a ...any) {
				cmd.PrintErrf(format, a...)
			},
		})
		if err != nil {
			cmd.PrintErrf("Error building site: %v\n", err)
			exitFunc(1)
			return
		}
		cmd.Println("Build complete!")
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildOut, "out", "o", "public", "Output directory for the compiled site")
	buildCmd.Flags().BoolVar(&buildOffline, "offline", false, "Build site without downloading external vendor assets (KaTeX, Mermaid)")
	buildCmd.Flags().StringVar(&buildVendorDir, "vendor-dir", "", "Custom directory containing vendor assets (katex, mermaid)")
	rootCmd.AddCommand(buildCmd)
}
