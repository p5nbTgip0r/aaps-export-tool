package cmd

import (
	"aaps-export-tool/util"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	FormatForce   bool
	FormatConsole bool
	FormatOutput  string
)

// formatCmd represents the format command
var formatCmd = &cobra.Command{
	Use:   "format <file>",
	Short: "Converts preferences in an unencrypted settings export between a string and JSON object",
	Long: `Converts the 'content' key (aka preferences) in an unencrypted settings export between a string and JSON object.

AAPS cannot import a settings export when the 'content' key is not a string, but JSON as a string is hard to edit manually.
This command allows you to convert the preferences between JSON object and string, to allow for manual editing and re-importing.`,
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
		if FormatOutput != "" {
			outputPath = FormatOutput
		}

		if !FormatForce && util.IsEncrypted(data) {
			fmt.Println("Cannot format: input file is encrypted")
			return
		}

		var outputData []byte
		var convertedType string
		if util.IsPreferencesObject(data) {
			outputData = util.ConvertPreferencesToString(data)
			convertedType = "string"
		} else {
			outputData = util.ConvertPreferencesToObject(data)
			convertedType = "JSON object"
		}

		err = os.WriteFile(outputPath, outputData, 0644)
		if err != nil {
			panic(err)
		}

		absolutePath, _ := filepath.Abs(outputPath)
		if FormatOutput != "" {
			fmt.Printf("Converted preferences to %s and wrote to \"%s\" successfully\n", convertedType, absolutePath)
		} else {
			fmt.Printf("Converted preferences to %s successfully\n", convertedType)
		}
	},
}

func init() {
	rootCmd.AddCommand(formatCmd)

	formatCmd.Flags().BoolVarP(&FormatForce, "force", "f", false, "Don't check if the input is decrypted before converting")

	formatCmd.Flags().BoolVarP(&FormatConsole, "console", "c", false, "Write converted file to stdout")
	formatCmd.Flags().StringVarP(&FormatOutput, "out", "o", "", "Write output to the specified file (default: original file)")
	formatCmd.MarkFlagsMutuallyExclusive("console", "out")
}
