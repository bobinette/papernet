package papernet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTests(t *testing.T) map[string]struct {
	file   *os.File
	papers []chromeBookmarkPaper
} {
	dirname := path.Join("testfiles", "chrome")
	files, err := ioutil.ReadDir(dirname)
	require.NoError(t, err)

	tts := make(map[string]struct {
		file   *os.File
		papers []chromeBookmarkPaper
	})
	for _, f := range files {
		ext := path.Ext(f.Name())
		if ext != ".html" {
			continue
		}

		base := strings.TrimSuffix(f.Name(), ext)
		tt := tts[base] // Just to get a default struct to fill

		// Set file
		htmlFile := path.Join(dirname, f.Name())
		tt.file, err = os.Open(htmlFile)
		require.NoError(t, err, "could not open file %s", htmlFile)

		// Load expected papers
		jsonFile := path.Join(dirname, fmt.Sprintf("%s.json", base))
		file, err := os.Open(jsonFile)
		require.NoError(t, err, "could not open file %s", jsonFile)
		var papers []chromeBookmarkPaper
		err = json.NewDecoder(file).Decode(&papers)
		require.NoError(t, err, "could not load expected papers for %s", base)
		tt.papers = papers

		tts[base] = tt
	}

	return tts
}

func TestImportChromeBookmarks(t *testing.T) {
	tts := loadTests(t)

	for name, tt := range tts {
		papers, err := importChromeBookmarks(tt.file)
		assert.NoError(t, err, name)
		assert.Equal(t, tt.papers, papers, name)
	}
}
