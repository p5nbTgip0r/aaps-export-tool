package cmd

import (
	"aaps-export-tool/core"
	"github.com/AlecAivazis/survey/v2"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aaps-export-tool",
	Short: "A CLI tool for exported AndroidAPS settings files",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aaps-export-tool.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().BoolVarP(&core.Verbose, "verbose", "v", false, "Enable additional logging output")
}

func displayPasswordPrompt() (string, error) {
	password := ""
	prompt := &survey.Password{
		Message: "Enter your master password:",
	}
	err := survey.AskOne(prompt, &password)
	if err != nil {
		return "", err
	}

	return password, nil
}
