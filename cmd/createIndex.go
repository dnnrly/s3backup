package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dnnrly/s3backup/filemeta"
	"github.com/spf13/cobra"
)

// createIndexCmd represents the createIndex command
var createIndexCmd = &cobra.Command{
	Use:   "create-index",
	Short: "Generates an index file for known files in a location",
	Long: `This command generates an index file that sits at the root of
your S3 bucket, avoiding all of the index performance issues with
scanning all the files. It identifies all of the meta data you
need to manage the files that have been backed up.`,
	Run: func(cmd *cobra.Command, args []string) {
		index, err := filemeta.NewIndexFromRoot("", optIndexDirectory, filemeta.FilePathWalker, filemeta.FileHasher)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read files for index: %s", err)
		}

		data, err := index.Encode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create index: %s", err)
			os.Exit(1)
		}

		err = ioutil.WriteFile(optIndexFile, []byte(data), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to write index file: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createIndexCmd)

	createIndexCmd.Flags().StringVar(&optIndexFile, "file", optIndexFile, "Location of the index file to write")
}
