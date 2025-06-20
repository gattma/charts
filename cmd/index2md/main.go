// from https://github.com/onosproject/build-tools/tree/master/build/cmd/index2md

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/coreos/go-semver/semver"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed yamlAppsTemplate.md
	yamlAppsTemplateMarkdown string

	//go:embed yamlAppsTemplate.html
	yamlAppsTemplateHTML string
)

// Chart :
type Chart struct {
	APIVersion  string   `yaml:"apiVersion"`
	AppVersion  string   `yaml:"appVersion"`
	Version     string   `yaml:"version"`
	Created     string   `yaml:"created"`
	Description string   `yaml:"description"`
	Urls        []string `yaml:"urls"`
}

// IndexYaml :
type IndexYaml struct {
	Title      string             `yaml:"-"`
	APIVersion string             `yaml:"apiVersion"`
	Entries    map[string][]Chart `yaml:"entries"`
	Generated  string             `yaml:"generated"`
}

/**
 * A simple application that takes the generated index.yaml and outputs it in
 * a Markdown format - usually we pipe this to README.md when in the gh-pages branch
 */
func main() {
	file := flag.String("file", "docs/index.yaml", "name of YAML file to parse")
	title := flag.String("title", `bakito <img src="https://helm.sh/img/helm.svg" alt="Helm" style="width:32px;"/> Chart Releases`, "title for the output")
	htmlout := flag.Bool("html", false, "output HTML instead of Markdown")
	flag.Parse()
	indexYaml, err := getIndexYaml(*file, *title)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load %s.yaml %v\n", *file, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

	funcs := template.FuncMap{
		"isLast": func(list []Chart, i int) bool {
			return len(list)-1 == i
		},
		"fileName": func(link string) string {
			return path.Base(link)
		},
	}

	var tmplAppsList *template.Template
	if !*htmlout {
		tmplAppsList, err = template.New("yamlAppsTemplateMarkdown").Funcs(funcs).Parse(yamlAppsTemplateMarkdown)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to parse template yamlAppsTemplateMarkdown: %v", err)
			os.Exit(1)
		}
	} else {
		tmplAppsList, err = template.New("yamlAppsTemplateHtml").Funcs(funcs).Parse(yamlAppsTemplateHTML)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to parse template yamlAppsTemplateHtml: %v", err)
			os.Exit(1)
		}
	}

	err = tmplAppsList.Execute(os.Stdout, indexYaml)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to execute %v\n", err)
		os.Exit(1)
	}
}

func getIndexYaml(location string, title string) (IndexYaml, error) {
	indexYaml := &IndexYaml{}
	b, err := os.ReadFile(location)
	if err != nil {
		return IndexYaml{}, err
	}

	if err := yaml.Unmarshal(b, indexYaml); err != nil {
		return IndexYaml{}, err
	}

	indexYaml.Title = title

	for key := range indexYaml.Entries {
		sort.Slice(indexYaml.Entries[key], func(i, j int) bool {
			v1, err1 := semver.NewVersion(strings.ReplaceAll(indexYaml.Entries[key][i].Version, "v", ""))
			if err1 != nil {
				return true
			}
			v2, err2 := semver.NewVersion(strings.ReplaceAll(indexYaml.Entries[key][j].Version, "v", ""))
			if err2 != nil {
				return false
			}
			return v2.LessThan(*v1)
		})
	}

	return *indexYaml, nil
}