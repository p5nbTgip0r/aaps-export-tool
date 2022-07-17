package cmd

import (
	"aaps-export-tool/util"
	"encoding/hex"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	EncryptForce    bool
	EncryptConsole  bool
	EncryptOutput   string
	EncryptPassword string
	EncryptSalt     string
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt <file>",
	Short: "Encrypts an unencrypted AAPS settings export and outputs to a file",
	Long: `Encrypts the preferences of an unencrypted AAPS settings export and outputs to a file.

Examples:
aaps-export-tool encrypt export.json
aaps-export-tool encrypt export.json --out "encrypted.json"
aaps-export-tool encrypt export.json --console`,
	Args: pathArg,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			panic(err)
		}

		if !EncryptForce && util.IsEncrypted(data) {
			fmt.Println("Cannot encrypt: input file is already encrypted")
			return
		}

		password := EncryptPassword
		if EncryptPassword == "" {
			var err error
			password, err = displayPasswordPrompt()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		salt, _ := hex.DecodeString(EncryptSalt)
		if EncryptSalt == "" {
			salt, _ = util.GenerateSalt()
		}

		content := []byte(gjson.GetBytes(data, "content").String())
		encrypted, err := util.Encrypt([]byte(password), salt, content)
		if err != nil {
			panic(err)
		}

		outputData := util.ConvertToEncryptedFormat(data, salt, encrypted, util.Sha256(content))

		if EncryptConsole {
			fmt.Print(string(outputData))
			return
		}

		outputPath := EncryptOutput
		if outputPath == "" {
			ext := filepath.Ext(args[0])
			name := strings.TrimSuffix(args[0], ext)
			name = name + "_encrypted" + ext

			outputPath = name
		}

		err = os.WriteFile(outputPath, outputData, 0644)
		if err != nil {
			panic(err)
		}

		absolutePath, _ := filepath.Abs(outputPath)
		fmt.Printf("Encrypted settings were exported to \"%s\"\n", absolutePath)
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.Flags().BoolVarP(&EncryptConsole, "console", "c", false, "Write export to stdout")
	encryptCmd.Flags().StringVarP(&EncryptOutput, "out", "o", "", "Write export to the specified file (default: original filename with '_encrypted' before file extension)")
	encryptCmd.MarkFlagsMutuallyExclusive("console", "out")

	encryptCmd.Flags().BoolVarP(&EncryptForce, "force", "f", false, "Don't check if the input is unencrypted before encrypting")
	encryptCmd.Flags().StringVarP(&EncryptSalt, "salt", "s", "", "Manually specify the salt to be used in encryption")
	encryptCmd.Flags().StringVarP(&EncryptPassword, "password", "p", "", "Manually specify encryption password (only use if necessary, like in shell scripts)")
}
