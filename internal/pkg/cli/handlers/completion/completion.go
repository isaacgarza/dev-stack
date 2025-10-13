package completion

import (
	"context"
	"fmt"
	"os"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	pkgTypes "github.com/isaacgarza/dev-stack/internal/pkg/types"
	"github.com/spf13/cobra"
)

type CompletionHandler struct{}

func NewCompletionHandler() *CompletionHandler {
	return &CompletionHandler{}
}

func (h *CompletionHandler) ValidateArgs(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("completion requires exactly one argument (%s)",
			fmt.Sprintf("%v", pkgTypes.AllShellTypeStrings()))
	}

	shell := pkgTypes.ShellType(args[0])
	if !shell.IsValid() {
		return fmt.Errorf("unsupported shell: %s (supported: %v)",
			args[0], pkgTypes.AllShellTypeStrings())
	}

	return nil
}

func (h *CompletionHandler) GetRequiredFlags() []string {
	return []string{}
}

func (h *CompletionHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	shell := pkgTypes.ShellType(args[0])

	// Get the root command to generate completion for
	rootCmd := cmd.Root()

	switch shell {
	case pkgTypes.ShellTypeBash:
		return rootCmd.GenBashCompletion(os.Stdout)
	case pkgTypes.ShellTypeZsh:
		return rootCmd.GenZshCompletion(os.Stdout)
	case pkgTypes.ShellTypeFish:
		return rootCmd.GenFishCompletion(os.Stdout, true)
	case pkgTypes.ShellTypePowerShell:
		return rootCmd.GenPowerShellCompletion(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}
