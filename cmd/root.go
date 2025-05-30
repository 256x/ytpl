// cmd/root.go
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ytpl/internal/config"
	"ytpl/internal/player"
	"ytpl/internal/playlist"
	"ytpl/internal/state"

	"github.com/spf13/cobra"
)

var (
	cfg      *config.Config
	appState *state.PlayerState
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ytpl",
	Short: "ytpl is a CLI YouTube music player",
	Long: `ytpl is a command-line application designed for searching and playing YouTube videos.
It allows you to manage local music stock and create custom playlists.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add subcommands
	rootCmd.AddCommand(playCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(volCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(delCmd)
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(shuffleCmd) // Main shuffle command (for all stocked songs)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(prevCmd)

	// List command and its subcommands
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listMakeCmd)
	listCmd.AddCommand(listAddCmd)
	listCmd.AddCommand(listRemoveCmd)
	listCmd.AddCommand(listDelCmd)
	listCmd.AddCommand(listShowCmd)
	listCmd.AddCommand(listPlayCmd)
	listCmd.AddCommand(listShuffleCmd) // NEW: Subcommand for shuffling a specific playlist

	// Setup signal handling for graceful shutdown (e.g., Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Received interrupt signal. Shutting down ytpl...")
		if appState != nil && appState.PID != 0 {
			if err := player.StopPlayer(appState); err != nil {
				log.Printf("Error stopping player on shutdown: %v", err)
			}
		}
		os.Exit(0)
	}()

	// Configure log output to a file to avoid interfering with TUI.
	logFilePath := filepath.Join(os.TempDir(), "ytpl_debug.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		log.Printf("--- ytpl starting: %s ---\n", time.Now().Format(time.RFC3339))
		log.Printf("Log file: %s\n", logFilePath)
	} else {
		fmt.Fprintf(os.Stderr, "WARNING: Failed to open log file %s: %v. Logging to stderr.\n", logFilePath, err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Config file not found. Creating a default config.")
			defaultConfigPath, pathErr := config.GetConfigPath()
			if pathErr != nil {
				return fmt.Errorf("failed to determine default config path: %w", pathErr)
			}
			log.Printf("Creating a default config at %s\n", defaultConfigPath)
			if writeErr := os.MkdirAll(filepath.Dir(defaultConfigPath), 0755); writeErr != nil {
				return fmt.Errorf("failed to create config directory: %w", writeErr)
			}
			if writeErr := os.WriteFile(defaultConfigPath, []byte(config.GetDefaultConfigContent()), 0644); writeErr != nil {
				return fmt.Errorf("failed to write default config file: %w", writeErr)
			}
			cfg, err = config.LoadConfig() // Try loading again
			if err != nil {
				return fmt.Errorf("failed to load config after creating default: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	appState, err = state.LoadState(cfg)
	if err != nil {
		log.Printf("Error loading application state: %v. Starting with empty state.", err)
		appState = &state.PlayerState{
			Volume:        cfg.DefaultVolume,
			IPCSocketPath: cfg.PlayerIPCSocketPath,
		}
	}

	playlist.Init(cfg.PlaylistDir)

	return nil
}
