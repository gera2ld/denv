package cli

import (
	"denv/internal/env_manager"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/spf13/cobra"
)

func NewRootCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denv",
		Short: "DEnv CLI",
	}

	cmd.AddCommand(newRunCommand(envManager))
	cmd.AddCommand(newDeleteCommand(envManager))
	cmd.AddCommand(newImportCommand(envManager))
	cmd.AddCommand(newExportCommand(envManager))
	cmd.AddCommand(newKeysCommand(envManager))
	cmd.AddCommand(newRenameCommand(envManager))
	cmd.AddCommand(newEditCommand(envManager))
	cmd.AddCommand(newRecipientsCommand(envManager))
	cmd.AddCommand(newRecipientAddCommand(envManager))
	cmd.AddCommand(newRecipientDelCommand(envManager))
	cmd.AddCommand(newReindexCommand(envManager))
	cmd.AddCommand(newReencryptAllCommand(envManager))
	cmd.AddCommand(newCatCommand(envManager))

	return cmd
}

func newRunCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	var envKeys []string
	var export bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run command with environment variables",
		Long: `Run a command with environment variables loaded from the specified keys.
You can also export the environment variables to stdout using the --export flag.`,
		Args: cobra.ArbitraryArgs, // Accepts any arguments after the command
		RunE: func(cmd *cobra.Command, args []string) error {
			envVars := envManager.GetEnvs(envKeys)

			if export {
				for key, value := range envVars {
					fmt.Printf("%s=%s\n", key, value)
				}
				return nil
			}

			if len(args) == 0 {
				return errors.New("no command provided to run")
			}

			env := os.Environ()
			for key, value := range envVars {
				env = append(env, fmt.Sprintf("%s=%s", key, value))
			}

			command := args[0]
			commandArgs := args[1:]
			cmdExec := exec.Command(command, commandArgs...)
			cmdExec.Env = env
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Stdin = os.Stdin

			return cmdExec.Run()
		},
	}

	cmd.Flags().StringArrayVarP(&envKeys, "env", "e", []string{}, "Keys to load environment variables")
	cmd.Flags().BoolVar(&export, "export", false, "Print environment variables to stdout")

	return cmd
}

func newDeleteCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			err := envManager.DeleteEnv(key)
			return err
		},
	}
}

func newImportCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "import <source>",
		Short: "Import data from a directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]

			keys, err := envManager.ImportTree(source, "")
			if err != nil {
				return fmt.Errorf("failed to import data from %s: %w", source, err)
			}

			fmt.Printf("Successfully imported %d keys from %s\n", len(keys), source)
			return nil
		},
	}
}

func newExportCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	var outDir string
	var prefix string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all data to a directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use the provided output directory and prefix
			if outDir == "" {
				outDir = "env-data" // Default output directory
			}

			keys, err := envManager.ExportTree(outDir, prefix)
			if err != nil {
				return fmt.Errorf("failed to export data: %w", err)
			}

			fmt.Printf("Exported %d keys to %s\n", len(keys), outDir)
			return nil
		},
	}

	// Add flags for the output directory and prefix
	cmd.Flags().StringVarP(&outDir, "outDir", "o", "env-data", "Output directory")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Filter the keys by the prefix and strip it when writing to files")

	return cmd
}

func newKeysCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "keys",
		Short: "List all keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := envManager.ListEnvs()
			if err != nil {
				return err
			}
			for _, key := range keys {
				fmt.Println(key)
			}
			return nil
		},
	}
}

func newRenameCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <key> <newName>",
		Short: "Rename a key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			newName := args[1]
			parsed, err := envManager.GetEnv(key)
			if err != nil {
				return err
			}
			return envManager.SetEnv(newName, parsed)
		},
	}
}

func sanitizeKeyForFilename(key string) string {
	re := regexp.MustCompile(`[^\w.-]`)
	return re.ReplaceAllString(key, "_")
}

func newEditCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <key>",
		Short: "Edit the value of a key with $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			editor := os.Getenv("EDITOR")
			if editor == "" {
				return errors.New("$EDITOR is not set")
			}

			parsed, _ := envManager.GetEnv(key)

			oldValue := ""
			if parsed != nil {
				value, err := envManager.FormatValue(parsed, false)
				if err != nil {
					return fmt.Errorf("failed to format value: %w", err)
				}
				oldValue = value
			}

			safeKey := sanitizeKeyForFilename(key)
			tempFile, err := os.CreateTemp("", fmt.Sprintf("%s.yml", safeKey))
			if err != nil {
				return fmt.Errorf("failed to create temporary file: %w", err)
			}
			defer os.Remove(tempFile.Name())

			if _, err := tempFile.WriteString(oldValue); err != nil {
				return fmt.Errorf("failed to write to temporary file: %w", err)
			}
			tempFile.Close()

			cmdExec := exec.Command(editor, tempFile.Name())
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Stdin = os.Stdin

			if err := cmdExec.Run(); err != nil {
				return fmt.Errorf("failed to open editor: %w", err)
			}

			newValueBytes, err := os.ReadFile(tempFile.Name())
			if err != nil {
				return fmt.Errorf("failed to read temporary file: %w", err)
			}
			newValue := string(newValueBytes)

			if newValue == "" || newValue == oldValue {
				fmt.Println("No changes made.")
				return nil
			}

			parsed, err = envManager.ParseRawValue(newValue, false)
			if err != nil {
				return fmt.Errorf("failed to parse new value: %w", err)
			}

			if err := envManager.SetEnv(key, parsed); err != nil {
				return fmt.Errorf("failed to save updated value: %w", err)
			}

			fmt.Println("Updated key:", key)
			return nil
		},
	}
}

func newRecipientsCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "recipients",
		Short: "List all recipients",
		RunE: func(cmd *cobra.Command, args []string) error {
			recipients := envManager.UserConfig.Data.Recipients
			for _, recipient := range recipients {
				fmt.Println(recipient)
			}
			return nil
		},
	}
}

func newRecipientAddCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "recipientAdd <recipient>",
		Short: "Add a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			recipient := args[0]
			return envManager.UserConfig.AddRecipient(recipient)
		},
	}
}

func newRecipientDelCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "recipientDel <recipient>",
		Short: "Remove a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			recipient := args[0]
			return envManager.UserConfig.RemoveRecipient(recipient)
		},
	}
}

func newReindexCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "reindex",
		Short: "Rebuild index for all data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return envManager.BuildIndex()
		},
	}
}

func newReencryptAllCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "reencryptAll",
		Short: "Reencrypt all data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return envManager.ReencryptAll()
		},
	}
}

func newCatCommand(envManager *env_manager.DynamicEnvManager) *cobra.Command {
	return &cobra.Command{
		Use:   "cat <key>",
		Short: "Show the value of a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			parsed, err := envManager.GetEnv(key)
			if err != nil {
				return fmt.Errorf("failed to retrieve key: %w", err)
			}

			value, err := envManager.FormatValue(parsed, false)
			if err != nil {
				return fmt.Errorf("failed to format value: %w", err)
			}

			fmt.Println(value)
			return nil
		},
	}
}
