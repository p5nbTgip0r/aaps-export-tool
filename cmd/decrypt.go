package cmd

import (
	"aaps-export-tool/util"
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	DecryptForce             bool
	DecryptPreferencesObject bool
	DecryptOnlyPreferences   bool
	DecryptConsole           bool
	DecryptOutput            string
	DecryptPassword          string
)

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt <file>",
	Short: "Decrypts an AAPS settings export and outputs to a file",
	Long: `Decrypts the preferences of an AAPS settings export and outputs to a file.

Examples:
aaps-export-tool decrypt export.json
aaps-export-tool decrypt export.json --out "decrypted.json"
aaps-export-tool decrypt export.json --console
aaps-export-tool decrypt export.json --preferences-object
`,
	Args: pathArg,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			panic(err)
		}

		if !DecryptForce && !util.IsEncrypted(data) {
			fmt.Println("Cannot decrypt: input file is already decrypted")
			return
		}

		password := DecryptPassword
		if DecryptPassword == "" {
			var err error
			password, err = displayPasswordPrompt()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		salt, _ := hex.DecodeString(gjson.GetBytes(data, "security.salt").String())

		decrypted, err := util.Decrypt([]byte(password), salt, gjson.GetBytes(data, "content").String())
		if err != nil {
			panic(err)
		}

		if DecryptConsole {
			fmt.Print(string(decrypted))
			return
		}

		outputPath := DecryptOutput
		if outputPath == "" {
			ext := filepath.Ext(args[0])
			name := strings.TrimSuffix(args[0], ext)
			name = name + "_decrypted" + ext

			outputPath = name
		}

		outputData := decrypted

		if !DecryptOnlyPreferences {
			outputData = util.ConvertToUnencryptedFormat(data, decrypted)
		}

		if DecryptPreferencesObject {
			outputData = util.ConvertPreferencesToObject(outputData)
		}

		err = os.WriteFile(outputPath, outputData, 0644)
		if err != nil {
			panic(err)
		}

		absolutePath, _ := filepath.Abs(outputPath)
		fmt.Printf("Decrypted settings were exported to \"%s\"\n", absolutePath)
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().StringVarP(&DecryptPassword, "password", "p", "", "Manually specify encryption password (only use if necessary, like in shell scripts)")
	decryptCmd.Flags().BoolVarP(&DecryptForce, "force", "f", false, "Don't check if the input is encrypted before decrypting")

	decryptCmd.Flags().BoolVarP(&DecryptPreferencesObject, "preferences-object", "m", false, `Convert 'content' to a JSON object instead of string.
This flag is effectively a shortcut for running the 'format' command on the exported file.

This makes it easier to edit the preferences manually, but makes the decrypted output incompatible with AAPS. You can 
use the 'format' command to restore AAPS compatibility by converting the preferences between object and string storage.`)
	decryptCmd.Flags().BoolVar(&DecryptOnlyPreferences, "only-preferences", false, "Only export the decrypted preferences portion of the file")
	decryptCmd.MarkFlagsMutuallyExclusive("preferences-object", "only-preferences")

	decryptCmd.Flags().BoolVarP(&DecryptConsole, "console", "c", false, "Write decrypted export to stdout")
	decryptCmd.Flags().StringVarP(&DecryptOutput, "out", "o", "", "Write decrypted output to the specified file (default: original filename with '_decrypted' before file extension)")
	decryptCmd.MarkFlagsMutuallyExclusive("console", "out")
}
