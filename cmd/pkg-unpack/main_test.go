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
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/test"
)

var (
	errParse = `couldn't get version/kind; json parse error: json: cannot unmarshal string into Go value of type struct { APIVersion string "json:\"apiVersion,omitempty\""; Kind string "json:\"kind,omitempty\"" }`

	simple = []byte(`---
apiVersion: example.org/v1
kind: Example
metadata:
  name: example
spec:
  cool: true
`)

	stream = []byte(`---
apiVersion: example.org/v1
kind: Example
metadata:
  name: example
spec:
  cool: true
---
apiVersion: example.org/v1
kind: Example
metadata:
  name: another-example
spec:
  cool: true
`)

	typed = []byte(`---
apiVersion: example.org/v1beta1
kind: Example
`)

	invalid = []byte(`wat`)
)

func TestCopy(t *testing.T) {
	type want struct {
		dst []byte
		err error
	}
	cases := map[string]struct {
		reason string
		src    []byte
		want   want
	}{
		"OneDocument": {
			reason: "It should be possible to copy a single valid YAML document",
			src:    simple,
			want: want{
				dst: simple,
			},
		},
		"TwoDocuments": {
			reason: "It should be possible to copy a stream of YAML documents",
			src:    stream,
			want: want{
				dst: stream,
			},
		},
		"TypeOnly": {
			reason: "A YAML document is valid as long as it contains an apiVersion and kind",
			src:    typed,
			want: want{
				dst: typed,
			},
		},
		"Invalid": {
			reason: "A YAML document must contain an apiVersion and kind",
			src:    invalid,
			want: want{
				err: errors.Wrap(errors.New(errParse), "cannot parse YAML document"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			src := bytes.NewReader(tc.src)
			dst := &bytes.Buffer{}
			err := Copy(dst, src)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nCopy(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			got, err := ioutil.ReadAll(dst)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(string(tc.want.dst), string(got)); diff != "" {
				t.Errorf("\n%s\nCopy(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}

}
