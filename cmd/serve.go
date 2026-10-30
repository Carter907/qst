package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Carter907/qst/internal/ssg"
	"github.com/spf13/cobra"
)

var (
	servePort      int
	serveHost      string
	serveOpen      bool
	serveOffline   bool
	serveVendorDir string
)

var serveCmd = &cobra.Command{
	Use:   "serve [directory|archive.kng]",
	Short: "Start a live-reloading local development server for the knowledge graph",
	Long: `Start a local HTTP development server that compiles the knowledge graph into an isolated temporary workspace, watches for file changes, validates graph constraints in real time, and automatically reloads connected browsers via Server-Sent Events (SSE).

You can pass either a knowledge graph directory or a packaged .kng archive file.
If the requested port is already in use, serve automatically increments the port until an available port is found.`,
	Example: `# Serve the current directory on default port 8080
qst serve

# Serve a specific directory on port 3000 and open the browser
qst serve ./examples/Boolean_Logic --port 3000 --open

# Serve a packaged .kng archive directly in offline mode
qst serve bundle.kng --offline`,
	Args: cobra.MaximumNArgs(1),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		vendorDir := serveVendorDir
		if vendorDir == "" {
			vendorDir = os.Getenv("QST_VENDOR_DIR")
		}

		srv, err := ssg.NewDevServer(ssg.ServeOptions{
			TargetPath: target,
			Host:       serveHost,
			Port:       servePort,
			Open:       serveOpen,
			Offline:    serveOffline,
			VendorDir:  vendorDir,
			LogFunc: func(format string, a ...any) {
				cmd.Printf(format, a...)
			},
			ErrFunc: func(format string, a ...any) {
				cmd.PrintErrf(format, a...)
			},
		})
		if err != nil {
			return fmt.Errorf("Error starting server: %w", err)
		}

		parentCtx := cmd.Context()
		if parentCtx == nil {
			parentCtx = context.Background()
		}
		ctx, stop := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
		defer stop()

		if err := srv.Start(ctx); err != nil {
			return fmt.Errorf("Server error: %w", err)
		}
		return nil
	}),
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on (automatically probes next available port if busy)")
	serveCmd.Flags().StringVar(&serveHost, "host", "localhost", "Host address to bind to")
	serveCmd.Flags().BoolVarP(&serveOpen, "open", "O", false, "Open the default browser automatically on startup")
	serveCmd.Flags().BoolVar(&serveOffline, "offline", false, "Serve without downloading external vendor assets (KaTeX, Mermaid)")
	serveCmd.Flags().StringVar(&serveVendorDir, "vendor-dir", "", "Custom directory containing vendor assets (katex, mermaid)")
	rootCmd.AddCommand(serveCmd)
}
