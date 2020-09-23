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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/crossplane/crossplane-runtime/pkg/parser"
)

func main() {
	be := parser.NewFsBackend(afero.NewReadOnlyFs(afero.NewOsFs()), parser.FsDir("."), parser.FsFilters(parser.SkipNotYAML()))
	src, err := be.Init(context.Background())
	if err != nil {
		// NOTE(negz): fmt.Print prints to stderr.
		fmt.Print(err.Error())
		os.Exit(1)
	}
	defer src.Close()

	if err := Copy(os.Stdout, src); err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
}

// Copy to dst from src, returning an error if src is not a YAML
// stream containing only valid Kubernetes objects.
func Copy(dst io.Writer, src io.Reader) error {
	y := yaml.NewYAMLReader(bufio.NewReader(src))
	d := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		unstructuredscheme.NewUnstructuredCreator(),
		unstructuredscheme.NewUnstructuredObjectTyper(),
		json.SerializerOptions{Yaml: true})

	b := &bytes.Buffer{}
	for {
		doc, err := y.Read()
		if err != nil && err != io.EOF {
			return errors.New("cannot read YAML document")
		}
		if err == io.EOF {
			break
		}
		o, _, err := d.Decode(doc, nil, nil)
		if err != nil {
			return errors.Wrap(err, "cannot parse YAML document")
		}

		_, _ = b.Write([]byte("---\n")) // Writing to a buffer never errors.
		if err := d.Encode(o, b); err != nil {
			return errors.Wrap(err, "cannot encode YAML document")
		}
		if _, err := io.Copy(dst, b); err != nil {
			return errors.Wrap(err, "cannot output YAML document")
		}

		b.Reset()
	}
	return nil
}
