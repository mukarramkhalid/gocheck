package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	engine "github.com/gatariee/gocheck/engine"
	utils "github.com/gatariee/gocheck/utils"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "",
	Long:  ``,

	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		amsi, _ := cmd.Flags().GetBool("amsi")
		defender, _ := cmd.Flags().GetBool("defender")

		if !amsi && !defender {
			defender = true
		}

		var (
			defender_path string
			err           error
		)

		if defender {
			defender_path, err = FindDefenderPath("C:\\")
			if err != nil {
				utils.PrintErr(err.Error())
				return
			}
		} else {
			defender_path = ""
		}

		utils.PrintInfo(fmt.Sprintf("Found Windows Defender at %s", defender_path))

		token := engine.Scanner{
			File:       file,
			Amsi:       amsi,
			Defender:   defender,
			EnginePath: defender_path,
		}

		start := time.Now()
		engine.Run(token)
		elapsed := time.Since(start)

		utils.PrintOk(fmt.Sprintf("Total time elasped: %s", elapsed))
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringP("file", "f", "", "Binary to check")
	checkCmd.MarkFlagRequired("file")

	checkCmd.Flags().BoolP("amsi", "a", false, "Use AMSI to scan the binary")
	checkCmd.Flags().BoolP("defender", "d", false, "Use Windows Defender to scan the binary")
}

func GetFileSize(file string) (int64, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func FindDefenderPath(root string) (string, error) {
	/* We don't want to perform this expensive operation if we don't have to, so let's search common paths first! */
	paths := []string{
		"C:\\Program Files\\Windows Defender\\MpCmdRun.exe",
		"C:\\Program Files (x86)\\Windows Defender\\MpCmdRun.exe",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	/* Now, we can panic and search for MpCmdRun.exe */
	utils.PrintErr("Could not find Windows Defender in common paths, searching C:\\ recursively for MpCmdRun.exe...")

	var defenderPath string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {

			/* If error is due to permission, don't panic just yet */
			if os.IsPermission(err) {
				return nil
				/* Keep walking! :) */
			}

			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), "MpCmdRun.exe") {
			defenderPath = path
			return filepath.SkipDir
		}
		return nil
	})
	return defenderPath, err
}