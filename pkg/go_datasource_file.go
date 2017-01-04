/*
Copyright 2017 All rights reserved.
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

package pkg

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/helper/schema"
)

func goDataSourceFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFileRead,
		Schema: map[string]*schema.Schema{
			"template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Contents of the template you wish rendered",
			},
			"snippits": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The path to a directory containing snippits",
			},
			"vars": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     make(map[string]interface{}),
				Description: "A map of variables used within the template",
			},
			"rendered": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The rendered template",
			},
		},
	}
}

// dataSourceFileRead is responsible rendering the template content
func dataSourceFileRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := renderGoTemplate(d)
	if err != nil {
		return err
	}
	d.Set("rendered", rendered)
	d.SetId(hash(rendered))
	return nil
}

// renderGoTemplate is responsible for generating the template
func renderGoTemplate(d *schema.ResourceData) (string, error) {
	templateName := d.Get("template").(string)
	snippitsPath := d.Get("snippits").(string)
	vars := d.Get("vars").(map[string]interface{})

	// step: read in the template content or file
	content, _, err := pathorcontents.Read(templateName)
	if err != nil {
		return "", err
	}
	// step: load the main template
	tmpl, err := template.New("main").Funcs(templateFuncs()).Parse(content)
	if err != nil {
		return "", err
	}
	// step: load any snippits if required
	if snippitsPath != "" {
		// build a list of files under the directory
		var files []string
		dfiles, err := ioutil.ReadDir(snippitsPath)
		if err != nil {
			return "", err
		}
		for _, x := range dfiles {
			files = append(files, fmt.Sprintf("%s/%s", strings.TrimRight(snippitsPath, "/"), x.Name()))
		}

		// step: parse the snippit files and add to the template
		if len(files) > 0 {
			tmpl, err = tmpl.ParseFiles(files...)
			if err != nil {
				return "", err
			}
		}
	}

	// step: render the template
	rendered := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(rendered, "main", vars); err != nil {
		return "", err
	}

	return rendered.String(), nil
}

// templateFuncs is a list of templates methods we support
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"split": func(s, delim string) []string {
			return strings.Split(s, delim)
		},
		"join": func(s []string, sep string) string {
			return strings.Join(s, sep)
		},
		"empty": func(s string) bool {
			return s == ""
		},
	}
}

// hash is responsible for calculating the hash of a string
func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
