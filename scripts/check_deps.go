#!/usr/bin/env gorun

package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const depsFile = ".last_deps_update"

func main() {
	log.SetFlags(0)
	depsFi, err := os.Stat(depsFile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	newerThan := func(path string) bool {
		fi, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}
		return depsFi.ModTime().After(fi.ModTime())
	}

	if err == nil &&
		newerThan("package.json") &&
		newerThan(filepath.FromSlash("src/site/static/js/typings.json")) {
		return
	}

	cmd := exec.Command("npm", "install")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	if f, err := os.Create(depsFile); err != nil {
		log.Printf("Failed to record dependency update. Future runs of stage.sh will start more slowly than necessary. Error from os.Create: %v", err)
	} else {
		f.Close()
	}
}
