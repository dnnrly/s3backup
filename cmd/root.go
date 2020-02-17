package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"

	awss3 "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/spf13/cobra"

	"github.com/dnnrly/s3backup"
	"github.com/dnnrly/s3backup/s3"
)

var (
	cfgFile           = "config.yaml"
	optIndexDirectory = "."
	optIndexFile      = ".s3backup.yaml"
	verbose           = false

	indexFile = ".index.yaml"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "s3backup",
	Short: "Backup your files to S3",
	Long: `This too backs up your files to S3 so that you can have them in
the cloud. It will scan the location(s) that you specify and
attempt rudimentary de-duplication.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		s3backup.Verbose = verbose
	},
	Run: doUpload,
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
	rootCmd.Flags().StringVarP(&optIndexDirectory, "root", "r", optIndexDirectory, "index scan root directory")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, fmt.Sprintf("config file (default is %s)", cfgFile))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "Verbose output")
}

func doUpload(cmd *cobra.Command, args []string) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	
	log.Println("Reading config")
	config, err := s3backup.NewConfigFromFile(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	log.Println("Creating S3 resources")
	store, err := s3.NewStore(config.S3)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	log.Printf("Reading remote index from %s\n", config.S3.Bucket)
	remoteIndex := &s3backup.Index{}
	indexReader, err := store.GetByKey(indexFile)
	if err != nil {
		if aerr, ok := err.(awserr.Error) ;
			ok && aerr.Code() == awss3.ErrCodeNoSuchKey {
			fmt.Println("Remote index does not exist, using empty index")
			err = nil
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else {
		buf := &bytes.Buffer{}
		buf.ReadFrom(indexReader)
		remoteIndex, err = s3backup.NewIndex(buf.String())
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	log.Println("Creating index")
	localIndex, err := s3backup.NewIndexFromRoot(
		"backup",
		optIndexDirectory,
		s3backup.FilePathWalker,
		s3backup.FileHasher,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	diff := localIndex.Diff(remoteIndex)
	for path, srcFile := range diff.Files {
		r, err := os.Open(path)
		defer func() {
			r.Close()
		}()

		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		} else {
			localHash, err := s3backup.FileHasher(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}

			if localHash != srcFile.Hash {
				remoteIndex.Files[path] = srcFile
			}

			log.Printf("Uploading %s as %s\n", path, srcFile.Key)
			err = store.Save(srcFile.Key, r)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
	}

	r, err := remoteIndex.Encode()
	log.Printf("Uploading index as %s\n", indexFile)
	err = store.Save(indexFile, bytes.NewBufferString(r))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	log.Println("Finished")
	os.Exit(0)
}
