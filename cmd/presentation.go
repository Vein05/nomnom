package cmd

import (
	"fmt"
	"path/filepath"

	content "nomnom/internal/content"
	"nomnom/internal/utils"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

type cliPresenter struct{}

func newCLIPresenter() cliPresenter {
	return cliPresenter{}
}

func (cliPresenter) Infof(format string, args ...any) {
	fmt.Printf("%s %s\n", color.WhiteString("▶"), color.CyanString(format, args...))
}

func (cliPresenter) Warnf(format string, args ...any) {
	fmt.Printf("%s %s\n", color.WhiteString("▶"), color.YellowString(format, args...))
}

func (cliPresenter) Errorf(format string, args ...any) {
	fmt.Printf("%s %s\n", color.WhiteString("▶"), color.RedString(format, args...))
}

func (cliPresenter) Titlef(format string, args ...any) {
	fmt.Println(color.BlueString(format, args...))
}

func (cliPresenter) Divider() {
	fmt.Println(color.CyanString("══════════════════════"))
}

func (cliPresenter) Banner() {
	fmt.Println(color.BlueString(`
_  _                 _  _
| \| | ___  _ __    | \| | ___  _ __ ___
| .  |/ _ \| '_ \   | .  |/ _ \| '_ ' _ \
| |\  | (_) | | | |  | |\  | (_) | | | | | |
|_| \_|\___/|_| |_|  |_| \_|\___/|_| |_| |_|
`))
	fmt.Printf("%s %s\n", color.WhiteString("▶"), color.BlueString("https://github.com/vein05/nomnom"))
}

func (cliPresenter) Approve(action, oldName, newName string) (utils.ApprovalDecision, error) {
	prompt := promptui.Select{
		Label: fmt.Sprintf("Approve %s for %s to %s", action, oldName, newName),
		Items: []string{"yes", "no", "approve all"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return utils.ApprovalNo, err
	}

	switch result {
	case "yes":
		return utils.ApprovalYes, nil
	case "approve all":
		return utils.ApprovalAll, nil
	default:
		return utils.ApprovalNo, nil
	}
}

func (cliPresenter) PrintSummary(results []content.ProcessResult) {
	success := color.New(color.FgGreen).SprintFunc()
	failed := color.New(color.FgRed).SprintFunc()

	fmt.Println(color.CyanString("📊 Summary of Operations"))
	fmt.Println(color.CyanString("══════════════════════"))

	for _, result := range results {
		if result.Success {
			fmt.Printf("%s \033]8;;file://%s\033\\%s\033]8;;\033\\ → \033]8;;file://%s\033\\%s\033]8;;\033\\\n",
				success("✓"),
				result.OriginalPath,
				filepath.Base(result.OriginalPath),
				result.FullNewPath,
				filepath.Base(result.NewPath))
			continue
		}

		fmt.Printf("%s \033]8;;file://%s\033\\%s\033]8;;\033\\ (Error: %v)\n",
			failed("✗"),
			result.OriginalPath,
			filepath.Base(result.OriginalPath),
			result.Error)
	}
}

func (cliPresenter) PrintResults(results []content.ProcessResult, dryRun bool) int {
	successCount := 0

	fmt.Println(color.CyanString("📊 Processing Results"))
	fmt.Println(color.CyanString("══════════════════════"))

	for _, result := range results {
		if !result.Success {
			fmt.Printf("%s %s (Error: %v)\n",
				color.RedString("❌ Failed to process:"),
				filepath.Base(result.OriginalPath),
				result.Error)
			continue
		}

		successCount++
		if dryRun {
			fmt.Printf("%s \033]8;;file://%s\033\\%s\033]8;;\033\\ → %s\n",
				color.GreenString("🔍 Would rename:"),
				result.FullOriginalPath,
				filepath.Base(result.OriginalPath),
				filepath.Base(result.NewPath))
			continue
		}

		fmt.Printf("%s %s →\033]8;;file://%s\033\\%s\033]8;;\033\\ \n",
			color.GreenString("✅ Renamed:"),
			filepath.Base(result.OriginalPath),
			result.FullNewPath,
			filepath.Base(result.NewPath))
	}

	return successCount
}
