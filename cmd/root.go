package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dnnrly/s3backup/filemeta"
)

var (
	cfgFile           = "."
	optIndexDirectory = "."
	optIndexFile      = ".s3backup.yaml"
	verbose           = false
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "s3backup",
	Short: "Backup your files to S3",
	Long: `This too backs up your files to S3 so that you can have them in
the cloud. It will scan the location(s) that you specify and
attempt rudimentary de-duplication.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		filemeta.Verbose = verbose
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVarP(&optIndexDirectory, "root", "r", optIndexDirectory, "index scan root directory")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, fmt.Sprintf("config file (default is %s)", cfgFile))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "Verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}
