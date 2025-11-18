/*
Copyright 2021 Upbound Inc.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/crossplane/upjet/v2/pkg/pipeline"

	"github.com/crossplane-contrib/provider-upjet-github/config"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		panic("root directory is required to be given as argument")
	}
	rootDir := os.Args[1]
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		panic(fmt.Sprintf("cannot calculate the absolute path with %s", rootDir))
	}

	p, err := config.GetProvider(context.Background())
	if err != nil {
		panic("cannot get cluster provider")
	}

	pn, err := config.GetProviderNamespaced(context.Background())
	if err != nil {
		panic("cannot get namespaced provider")
	}

	pipeline.Run(p, pn, absRootDir)
}
