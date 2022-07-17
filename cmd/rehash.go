package cmd

import (
	"aaps-export-tool/util"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	RehashOutput string
)

// rehashCmd represents the rehash command
var rehashCmd = &cobra.Command{
	Use:   "rehash <file>",
	Short: "Re-calculates the file hash embedded in an export file",
	Long: `Re-calculates the 'file_hash' value in an export file.
This is useful when the export file has been modified manually and the file hash is no longer valid.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires an input file")
		}

		_, err := ioutil.ReadFile(args[0])
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			panic(err)
		}

		outputPath := args[0]
		if RehashOutput != "" {
			outputPath = RehashOutput
		}

		outputData := util.CalculateFileHash(data)

		err = os.WriteFile(outputPath, outputData, 0644)
		if err != nil {
			panic(err)
		}

		absolutePath, _ := filepath.Abs(outputPath)
		if RehashOutput != "" {
			fmt.Printf("Recalculated file hash and wrote to \"%s\" successfully\n", absolutePath)
		} else {
			fmt.Println("File hash was recalculated successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(rehashCmd)

	rehashCmd.Flags().StringVarP(&RehashOutput, "out", "o", "", "Write output to the specified file (default: original file)")
}
