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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/crossplane-runtime/pkg/parser"
)

func main() {
	if err := Run(context.Background()); err != nil {
		fmt.Print(err.Error())
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
	p := parser.New(uScheme, uScheme)
	b := parser.NewFsBackend(afero.NewReadOnlyFs(afero.NewOsFs()), parser.FsDir("."), parser.FsFilters(parser.SkipNotYAML()))
	reader, err := b.Init(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot initialize filesystem backend")
	}
	pkg, err := p.Parse(ctx, reader)
	if err != nil {
		return errors.Wrap(err, "cannot parse the files")
	}
	list := append(pkg.GetMeta(), pkg.GetObjects()...)
	for _, m := range list {
		if m.GetObjectKind().GroupVersionKind().Empty() {
			continue
		}
		u, ok := m.(*unstructured.Unstructured)
		if !ok {
			return errors.New("object cannot be casted into *unstructured.Unstructured")
		}
		out, err := yaml.Marshal(u.UnstructuredContent())
		if err != nil {
			return errors.Wrap(err, "cannot marshall meta object into yaml")
		}
		// Leaving the new line character to the OS instead of one fmt.Printf.
		fmt.Println("---")
		fmt.Print(string(out))
	}
	return nil
}
