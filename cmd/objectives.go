package cmd

import (
	"aaps-export-tool/core"
	"aaps-export-tool/util"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	ObjectivesList     []int
	ObjectivesConsole  bool
	ObjectivesOutput   string
	ObjectivesPassword string
)

// objectivesCmd represents the objectives command
var objectivesCmd = &cobra.Command{
	Use:   "objectives <file>",
	Short: "Edit completion state of objectives",
	Long: `Edits the completion state of objectives in a settings export.

Examples:
aaps-export-tool objectives export.json
aaps-export-tool objectives export.json --out "export-objectives.json"
aaps-export-tool objectives export.json -j 4 -j 5
aaps-export-tool objectives export.json -j 6,7,8`,
	Hidden: true,
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

		isPrefsObject := util.IsPreferencesObject(data)
		isEncrypted := util.IsEncrypted(data)
		password := ""

		if isEncrypted {
			// decrypt the data first
			password, err = displayPasswordPrompt()
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			salt, _ := hex.DecodeString(gjson.GetBytes(data, "security.salt").String())

			decrypted, err := util.Decrypt([]byte(password), salt, gjson.GetBytes(data, "content").String())
			if err != nil {
				panic(err)
			}
			data = util.ConvertToUnencryptedFormat(data, decrypted)
		}

		prefs := []byte(gjson.GetBytes(data, "content").String())

		if len(ObjectivesList) == 0 {
			defaults := util.GetCompletedObjectives(prefs)
			objs, err := selectObjectives(defaults)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			ObjectivesList = objs
		}

		if len(ObjectivesList) == 0 {
			fmt.Println("No objectives were selected")
			return
		}

		objList := util.ObjectiveNumbersToObjects(ObjectivesList)
		for _, obj := range objList {
			prefs = obj.Complete(prefs)
			if core.Verbose {
				fmt.Printf("Set objective %d (%s) as completed\n", obj.Number, obj.Name)
			}
		}

		if isEncrypted {
			// re-encrypt if original was encrypted
			salt, _ := util.GenerateSalt()
			encrypted, err := util.Encrypt([]byte(password), salt, prefs)
			if err != nil {
				panic(err)
			}

			data = util.ConvertToEncryptedFormat(data, salt, encrypted, util.Sha256(prefs))
		} else {
			data, _ = sjson.SetRawBytes(data, "content", prefs)

			// restore the previous storage format for the prefs
			if isPrefsObject {
				data = util.ConvertPreferencesToObject(data)
			} else {
				data = util.ConvertPreferencesToString(data)
			}
		}

		if ObjectivesConsole {
			fmt.Print(string(data))
			return
		}

		outputPath := ObjectivesOutput
		if outputPath == "" {
			ext := filepath.Ext(args[0])
			name := strings.TrimSuffix(args[0], ext)
			name = name + "_objectives" + ext

			outputPath = name
		}

		err = os.WriteFile(outputPath, data, 0644)
		if err != nil {
			panic(err)
		}

		vals, _ := json.Marshal(ObjectivesList)
		absolutePath, _ := filepath.Abs(outputPath)
		fmt.Printf("Objectives %s are now completed and the file was exported to \"%s\"\n", vals, absolutePath)
	},
}

func init() {
	rootCmd.AddCommand(objectivesCmd)

	objectivesCmd.Flags().StringVarP(&ObjectivesPassword, "password", "p", "", "Manually specify encryption password (only use if necessary, like in shell scripts)")
	objectivesCmd.Flags().IntSliceVarP(&ObjectivesList, "objectives", "j", []int{}, "Comma-separated objective number(s) to mark as completed. May be specified multiple times")

	objectivesCmd.Flags().BoolVarP(&ObjectivesConsole, "console", "c", false, "Write export to stdout")
	objectivesCmd.Flags().StringVarP(&ObjectivesOutput, "out", "o", "", "Write output to the specified file (default: original filename with '_objectives' before file extension)")
	objectivesCmd.MarkFlagsMutuallyExclusive("console", "out")
}

func selectObjectives(defaults []int) ([]int, error) {
	var selectedOptions []string

	optionsMap := make(map[string]int)
	optionsDisplay := make([]string, len(util.Objectives))
	defaultOptions := make([]string, len(defaults))
	for i, obj := range util.Objectives {
		display := fmt.Sprintf("Objective %d (%s)", i+1, obj.Name)

		optionsDisplay[i] = display
		optionsMap[display] = i + 1

		for i2, def := range defaults {
			if def == obj.Number {
				defaultOptions[i2] = display
			}
		}
	}

	prompt := &survey.MultiSelect{
		Message:  "Select objectives to mark as completed: (unselected ones will not be affected)",
		Options:  optionsDisplay,
		Default:  defaultOptions,
		PageSize: 10,
	}
	err := survey.AskOne(prompt, &selectedOptions)
	if err != nil {
		return nil, err
	}

	selectedObjectives := make([]int, len(selectedOptions))

	for i, optName := range selectedOptions {
		selectedObjectives[i] = optionsMap[optName]
	}

	return selectedObjectives, nil
}
