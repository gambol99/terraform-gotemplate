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
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var testProviders = map[string]terraform.ResourceProvider{
	"gotemplate": Provider(),
}

func TestGoDataSourceFile(t *testing.T) {
	resource := goDataSourceFile()
	if resource == nil {
		t.Error("we should have recieved the provider schema")
	}
}

func TestGoTemplate(t *testing.T) {
	cases := []struct {
		Content  string
		Vars     string
		Expected string
	}{
		{
			Vars:     "{}",
			Content:  "This is a template!",
			Expected: "This is a template!",
		},
		{
			Vars:     `{name="rohith"}`,
			Content:  "Hello {{ .name }}",
			Expected: "Hello rohith",
		},
		{
			Vars:     `{enabled=true}`,
			Content:  "{{ if is_true .enabled }}is_true{{ end }}",
			Expected: "is_true",
		},
		{
			Vars:     `{enabled=false}`,
			Content:  "{{ if is_true .enabled }}is_true{{ else }}is_false{{end}}",
			Expected: "is_false",
		},
	}

	for _, x := range cases {
		resource.UnitTest(t, resource.TestCase{
			Providers: testProviders,
			Steps: []resource.TestStep{
				{
					Config: testTemplateConfig(x.Content, x.Vars),
					Check: func(s *terraform.State) error {
						got := s.RootModule().Outputs["rendered"]
						if x.Expected != got.Value {
							return fmt.Errorf("template:\n%s\nvars:\n%s\ngot:\n%s\nwant:\n%s\n", x.Content, x.Vars, got, x.Expected)
						}
						return nil
					},
				},
			},
		})
	}
}

func testTemplateConfig(template, vars string) string {
	return fmt.Sprintf(`
		data "gotemplate_file" "test" {
			template = "%s"
			vars     = %s
		}
		output "rendered" {
			value = "${data.gotemplate_file.test.rendered}"
		}`, template, vars)
}
