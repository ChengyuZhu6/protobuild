/*
   Copyright The containerd Authors.

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
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var (
	tmpl = template.Must(template.New("protoc").Parse(`protoc -I
	{{- range $index, $include := .Includes -}}
		{{if $index}}` + string(filepath.ListSeparator) + `{{end -}}
			{{.}}
	{{- end -}}
	{{- if .Descriptors}} --include_imports --descriptor_set_out={{.Descriptors}}{{- end -}}

	{{- range $gen := .Generators }} --{{- $gen.Name -}}_out={{- $gen.OutputDir }}
		{{- range $k, $v := $gen.Parameters }} --{{- $gen.Name -}}_opt={{- $k -}}={{- $v -}}{{- end -}}
	{{- end -}}

	{{- range .Files}} {{.}}{{end -}}
`))
)

type generator struct {
	Name       string
	OutputDir  string
	Parameters map[string]string
}

// protocParams defines inputs to a protoc command string.
type protocCmd struct {
	Generators  []generator
	Includes    []string
	Descriptors string
	Files       []string
}

func (p *protocCmd) mkcmd() (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (p *protocCmd) run() error {
	arg, err := p.mkcmd()
	if err != nil {
		return err
	}

	// pass to sh -c so we don't need to re-split here.
	args := []string{shArg, arg}
	cmd := exec.Command(shCmd, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
