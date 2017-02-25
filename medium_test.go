package papernet

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"testing"
)

var mediumTemplate = `
<!DOCTYPE html><html><body><script>// <![CDATA[
var GLOBALS = {}
// ]]></script><script charset="UTF-8" src="trololo.cdn" async></script><script>// <![CDATA[
window["obvInit"](%s)
// ]]></script></body></html>
`

func TestMediumImporter_Import(t *testing.T) {
	type expected struct {
		Title   string
		Summary string
		Tags    []string
		Authors []string
	}
	tts := map[string]struct {
		Filename string
		Error    bool
		Expected expected
	}{
		"death star design": {
			Filename: "deathstar",
			Error:    false,
			Expected: expected{
				Title:   "How we designed the Death Star, and why we failed at protecting the plans",
				Summary: "In this document, I will share with you the process we went through when designing the Death Star. Moreover, I will also discuss how and why we got the plans stolen by the rebellion",
				Tags:    []string{"Design", "Star Wars"},
				Authors: []string{"Darth Vader"},
			},
		},
	}

	importer := MediumImporter{}
	for name, tt := range tts {
		data, err := ioutil.ReadFile(path.Join("testfiles", fmt.Sprintf("medium_%s.json", tt.Filename)))
		if err != nil {
			t.Errorf("%s - error reading file %s: %v", name, tt.Filename, err)
			continue
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, mediumTemplate, string(data))
		}))
		defer ts.Close()

		paper, err := importer.Import(ts.URL)
		if err != nil && !tt.Error {
			t.Errorf("%s - should not have failed but did with error %v", name, err)
		} else if err == nil && tt.Error {
			t.Errorf("%s - should have failed but did not", name)
		} else {
			if paper.Title != tt.Expected.Title {
				t.Errorf("%s - incorrect title: expected %s got %s", name, tt.Expected.Title, paper.Title)
			}

			if paper.Summary != tt.Expected.Summary {
				t.Errorf("%s - incorrect summary: expected %s got %s", name, tt.Expected.Summary, paper.Summary)
			}

			if !reflect.DeepEqual(paper.Tags, tt.Expected.Tags) {
				t.Errorf("%s - incorrect tags: expected %s got %s", name, tt.Expected.Tags, paper.Tags)
			}

			if !reflect.DeepEqual(paper.Authors, tt.Expected.Authors) {
				t.Errorf("%s - incorrect authors: expected %s got %s", name, tt.Expected.Authors, paper.Authors)
			}
		}
	}
}
