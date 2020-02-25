package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
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

func doLog(format string, args ...interface{}) {
	if verbose {
		log.Printf(format, args...)
	}
}

func doUpload(cmd *cobra.Command, args []string) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	config := readConfig()
	store := createStore(config.S3)
	remoteIndex := readRemoteIndex(config, store)
	localIndex := createLocalIndex()
	s3backup.UploadDifferences(localIndex, remoteIndex, store, getFile)
	uploadIndex(localIndex, store)

	doLog("Finished")
	os.Exit(0)
}

func readConfig() *s3backup.Config {
	doLog("Reading config")
	config, err := s3backup.NewConfigFromFile(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return config
}

func createStore(config s3.Config) *s3.Store {
	doLog("Creating S3 resources")
	store, err := s3.NewStore(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return store
}

func readRemoteIndex(config *s3backup.Config, store *s3.Store) *s3backup.Index {
	doLog("Reading remote index from %s\n", config.S3.Bucket)
	remoteIndex := &s3backup.Index{}
	indexReader, err := store.GetByKey(indexFile)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == awss3.ErrCodeNoSuchKey {
			doLog("Remote index does not exist, using empty index")
			err = nil
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else {
		buf := &bytes.Buffer{}
		_, _ = buf.ReadFrom(indexReader)
		remoteIndex, err = s3backup.NewIndex(buf.String())
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	return remoteIndex
}

func createLocalIndex() *s3backup.Index {
	doLog("Creating index")
	localIndex, err := s3backup.NewIndexFromRoot(
		"",
		optIndexDirectory,
		s3backup.FilePathWalker,
		s3backup.FileHasher,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return localIndex
}

func getFile(p string) io.ReadCloser {
	r, err := os.Open(path.Clean(p))

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return r
}

func uploadIndex(index *s3backup.Index, store *s3.Store) {
	r, _ := index.Encode()
	doLog("Uploading index as %s\n", indexFile)
	err := store.Save(indexFile, bytes.NewBufferString(r))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
