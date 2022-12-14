/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	ExitOK    int = 0
	ExitError int = 1
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitError
	}
	return ExitOK
}

func GofmtalMain(filename string, writer io.Writer) error {
	// formattedCode, err := processFile(filename)
	formattedCode, err := processFile(filename)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, formattedCode)
	if err != nil {
		return err
	}

	return nil
}

func runE(cmd *cobra.Command, args []string) error {
	// TODO: 自由に指定できるようにする
	var out io.Writer
	out = os.Stdout

	var errs []error

	for _, arg := range args {
		switch info, err := os.Stat(arg); {

		case err != nil:
			errs = append(errs, err)
			continue

		case !info.IsDir():
			err := GofmtalMain(arg, out)
			if err != nil {
				errs = append(errs, err)
				continue
			}

		default:
			// ディレクトリ下のすべてのファイルをfilesに追加する
			var files []string
			err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					files = append(files, path)
				}
				return err
			})
			if err != nil {
				errs = append(errs, err)
				continue
			}

			// TODO: 79行目と同じ処理なのでまとめたい
			for _, file := range files {
				// skip not gofile
				if !IsGoFile(file) {
					continue
				}
				err := GofmtalMain(file, out)
				if err != nil {
					errs = append(errs, err)
					continue
				}
			}
		}
	}
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gofmtal",
	Short: "gofmtal is extended source code functionality in comments to gofmt.",
	Long:  "",
	RunE:  runE,
}
