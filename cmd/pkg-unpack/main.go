/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/crossplane/crossplane-runtime/pkg/parser"
)

func main() {
	if err := Run(context.Background()); err != nil {
		// We want our YAML stream written to stdout, but errors written to
		// stderr. println prints to stderr.
		println(err.Error())
		os.Exit(1)
	}
}

func Run(ctx context.Context) error {
	uScheme := struct {
		runtime.ObjectTyper
		runtime.ObjectCreater
	}{
		ObjectTyper:   unstructuredscheme.NewUnstructuredObjectTyper(),
		ObjectCreater: unstructuredscheme.NewUnstructuredCreator(),
	}

	// We want to leverage the package parser, but we don't need to make a
	// distinction between package metadata and payload objects. Passing an
	// empty 'metadata' scheme and an unstructured 'object' scheme will cause
	// the parser to treat any valid Kubernetes YAML (including package metadata
	// objects) as regular payload objects.
	p := parser.New(runtime.NewScheme(), uScheme)
	b := parser.NewFsBackend(afero.NewReadOnlyFs(afero.NewOsFs()), parser.FsDir("."), parser.FsFilters(parser.SkipNotYAML()))
	reader, err := b.Init(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot initialize filesystem backend")
	}

	// Parse will load any files that end with .yaml and that appear to be
	// valid Kubernetes objects (e.g. have an apiVersion and kind) and expose
	// them via its GetObjects method. Files that end with .yaml but that are
	// not valid YAML, or that don't have an apiVersion and kind, will result in
	// an error being returned.
	pkg, err := p.Parse(ctx, reader)
	if err != nil {
		return errors.Wrap(err, "cannot parse the files")
	}

	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, uScheme, uScheme, json.SerializerOptions{Yaml: true})
	for _, m := range pkg.GetObjects() {
		fmt.Println("---") // fmt.Println writes to stdout.
		if err := s.Encode(m, os.Stdout); err != nil {
			return errors.Wrap(err, "cannot marshal the object into yaml")
		}
	}
	return nil
}
