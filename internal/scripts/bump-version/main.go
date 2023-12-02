package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/robgonnella/ops/internal/scripts/bump-version/version"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal(errors.New("must provide version as argument"))
	}

	versionStr := args[0]
	outFile := "internal/app-info/info.go"
	templatePath := "internal/templates/info.go.tmpl"

	git := version.NewGit()
	generator := version.NewTemplateGenerator(outFile, templatePath)

	execData := version.BumpData{
		Version:      versionStr,
		OutFile:      outFile,
		TemplatePath: templatePath,
	}

	if err := version.Bump(execData, generator, git); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully bumped version to %s\n", versionStr)

	fmt.Println("To deploy run: \"git push <repo> <branch> --tags\"")
}
