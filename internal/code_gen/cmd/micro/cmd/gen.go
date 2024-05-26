/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/markity/micro/internal/code_gen/util"

	"github.com/spf13/cobra"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "provides proto and generates code",
	Long:  `provides include path(if necessary) and a service proto path for code generation`,
	Run:   nil,
}

func init() {
	includesInput := genCmd.Flags().StringP("include", "I", "",
		"provide proto include path, such as \"micro gen -I ./idl ./idl/service/user.proto\"")
	genCmd.Run = func(cmd *cobra.Command, args []string) {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		_, _, ok := util.SearchGoMod(dir)
		if !ok {
			cmd.Println("go mod not found")
		}

		if len(args) != 1 {
			cmd.Printf("one param is expected\n")
			return
		}

		err = os.Mkdir("micro_gen", os.ModePerm)
		if err != nil {
			if !errors.Is(err, os.ErrExist) {
				panic(err)
			}
		}

		cmdArgs := []string{}
		if includesInput != nil {
			cmdArgs = append(cmdArgs, "-I", *includesInput)
		}

		protoPath := filepath.Join(dir, args[0])
		cmdArgs = append(cmdArgs, `--micro_out=./micro_gen`, fmt.Sprintf(`--micro_opt=gen_path=%s/micro_gen,proto_path=%s`, dir, protoPath), args[0])

		tobeRunCmd := exec.Command("protoc", cmdArgs...)
		tobeRunCmd.Stdout = os.Stdout
		tobeRunCmd.Stderr = os.Stderr
		err = tobeRunCmd.Run()
		if err != nil {
			cmd.Println("failed to generate code:", err)
			return
		}

		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Stdout = os.Stdout
		tidyCmd.Stderr = os.Stderr
		err = tidyCmd.Run()
		if err != nil {
			cmd.Println("go mod tidy failed")
			return
		}

		fmt.Println("ok.")
	}
	rootCmd.AddCommand(genCmd)
}
