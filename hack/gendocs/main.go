package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/appscode/go/runtime"
	"github.com/appscode/guard/commands"
	"github.com/spf13/cobra/doc"
)

const (
	version = "0.1.0-rc.5"
)

// ref: https://github.com/spf13/cobra/blob/master/doc/md_docs.md
func main() {
	genCLIDocs()
	genServerDocs()
}

func genCLIDocs() {
	var (
		tplFrontMatter = template.Must(template.New("index").Parse(`---
title: Guard CLI Reference
description: Guard CLI Reference
menu:
  product_guard_{{ .Version }}:
    identifier: guard-cli
    name: Guard CLI
    parent: reference
    weight: 20
menu_name: product_guard_{{ .Version }}
---
`))

		_ = template.Must(tplFrontMatter.New("cmd").Parse(`---
title: {{ .Name }}
menu:
  product_guard_{{ .Version }}:
    identifier: {{ .ID }}
    name: {{ .Name }}
    parent: guard-cli
{{- if .RootCmd }}
    weight: 0
{{ end }}
product_name: guard
section_menu_id: reference
menu_name: product_guard_{{ .Version }}
{{- if .RootCmd }}
url: /products/guard/{{ .Version }}/reference/guard-cli/
aliases:
  - products/guard/{{ .Version }}/reference/guard-cli/guard-cli/
{{ end }}
---
`))
	)
	rootCmd := commands.NewRootCmdCLI(version)
	dir := runtime.GOPath() + "/src/github.com/appscode/guard/docs/reference/guard-cli"
	fmt.Printf("Generating cli markdown tree in: %v\n", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	filePrepender := func(filename string) string {
		filename = filepath.Base(filename)
		base := strings.TrimSuffix(filename, path.Ext(filename))
		name := strings.Title(strings.Replace(base, "_", " ", -1))
		parts := strings.Split(name, " ")
		if len(parts) > 1 {
			name = strings.Join(parts[1:], " ")
		}
		data := struct {
			ID      string
			Name    string
			Version string
			RootCmd bool
		}{
			strings.Replace(base, "_", "-", -1),
			name,
			version,
			!strings.ContainsRune(base, '_'),
		}
		var buf bytes.Buffer
		if err := tplFrontMatter.ExecuteTemplate(&buf, "cmd", data); err != nil {
			log.Fatalln(err)
		}
		return buf.String()
	}

	linkHandler := func(name string) string {
		return "/docs/reference/guard-cli/" + name
	}
	doc.GenMarkdownTreeCustom(rootCmd, dir, filePrepender, linkHandler)

	index := filepath.Join(dir, "_index.md")
	f, err := os.OpenFile(index, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	err = tplFrontMatter.ExecuteTemplate(f, "index", struct{ Version string }{version})
	if err != nil {
		log.Fatalln(err)
	}
	if err := f.Close(); err != nil {
		log.Fatalln(err)
	}
}

func genServerDocs() {
	var (
		tplFrontMatter = template.Must(template.New("index").Parse(`---
title: Guard Server Reference
description: Guard Server Reference
menu:
  product_guard_{{ .Version }}:
    identifier: guard-server
    name: Searchlight
    parent: reference
    weight: 20
menu_name: product_guard_{{ .Version }}
---
`))

		_ = template.Must(tplFrontMatter.New("cmd").Parse(`---
title: {{ .Name }}
menu:
  product_guard_{{ .Version }}:
    identifier: {{ .ID }}
    name: {{ .Name }}
    parent: guard-server
{{- if .RootCmd }}
    weight: 0
{{ end }}
product_name: guard
section_menu_id: reference
menu_name: product_guard_{{ .Version }}
{{- if .RootCmd }}
url: /products/guard/{{ .Version }}/reference/guard-server/
aliases:
  - products/guard/{{ .Version }}/reference/guard-server/guard-server/
{{ end }}
---
`))
	)
	rootCmd := commands.NewRootCmdServer(version)
	dir := runtime.GOPath() + "/src/github.com/appscode/guard/docs/reference/guard-server"
	fmt.Printf("Generating cli markdown tree in: %v\n", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	filePrepender := func(filename string) string {
		filename = filepath.Base(filename)
		base := strings.TrimSuffix(filename, path.Ext(filename))
		name := strings.Title(strings.Replace(base, "_", " ", -1))
		parts := strings.Split(name, " ")
		if len(parts) > 1 {
			name = strings.Join(parts[1:], " ")
		}
		data := struct {
			ID      string
			Name    string
			Version string
			RootCmd bool
		}{
			strings.Replace(base, "_", "-", -1),
			name,
			version,
			!strings.ContainsRune(base, '_'),
		}
		var buf bytes.Buffer
		if err := tplFrontMatter.ExecuteTemplate(&buf, "cmd", data); err != nil {
			log.Fatalln(err)
		}
		return buf.String()
	}

	linkHandler := func(name string) string {
		return "/docs/reference/guard-server/" + name
	}
	doc.GenMarkdownTreeCustom(rootCmd, dir, filePrepender, linkHandler)

	index := filepath.Join(dir, "_index.md")
	f, err := os.OpenFile(index, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	err = tplFrontMatter.ExecuteTemplate(f, "index", struct{ Version string }{version})
	if err != nil {
		log.Fatalln(err)
	}
	if err := f.Close(); err != nil {
		log.Fatalln(err)
	}
}
