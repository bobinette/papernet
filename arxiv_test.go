package papernet

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var arxivResponse = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <link href="http://arxiv.org/api/query?search_query%3D%26id_list%3D1234.5678%26start%3D0%26max_results%3D10" rel="self" type="application/atom+xml"/>
  <title type="html">ArXiv Query: search_query=&amp;id_list=1234.5678&amp;start=0&amp;max_results=10</title>
  <id>http://arxiv.org/api/zNPcAyYNIo22QOQOKpV4i12np5Q</id>
  <updated>2017-02-23T00:00:00-05:00</updated>
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">10</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v5</id>
    <updated>2016-12-29T19:05:11Z</updated>
    <published>2015-12-08T04:46:38Z</published>
    <title>SSD: Single Shot MultiBox Detector</title>
    <summary>  We present a method for detecting objects in images using a single deep
neural network. Our approach, named SSD, discretizes the output space of
bounding boxes into a set of default boxes over different aspect ratios and
scales per feature map location. At prediction time, the network generates
scores for the presence of each object category in each default box and
produces adjustments to the box to better match the object shape. Additionally,
the network combines predictions from multiple feature maps with different
resolutions to naturally handle objects of various sizes. Our SSD model is
simple relative to methods that require object proposals because it completely
eliminates proposal generation and subsequent pixel or feature resampling stage
and encapsulates all computation in a single network. This makes SSD easy to
train and straightforward to integrate into systems that require a detection
component. Experimental results on the PASCAL VOC, MS COCO, and ILSVRC datasets
confirm that SSD has comparable accuracy to methods that utilize an additional
object proposal step and is much faster, while providing a unified framework
for both training and inference. Compared to other single stage methods, SSD
has much better accuracy, even with a smaller input image size. For $300\times
300$ input, SSD achieves 72.1% mAP on VOC2007 test at 58 FPS on a Nvidia Titan
X and for $500\times 500$ input, SSD achieves 75.1% mAP, outperforming a
comparable state of the art Faster R-CNN model. Code is available at
https://github.com/weiliu89/caffe/tree/ssd .
</summary>
    <author>
      <name>Wei Liu</name>
    </author>
    <author>
      <name>Dragomir Anguelov</name>
    </author>
    <author>
      <name>Dumitru Erhan</name>
    </author>
    <author>
      <name>Christian Szegedy</name>
    </author>
    <author>
      <name>Scott Reed</name>
    </author>
    <author>
      <name>Cheng-Yang Fu</name>
    </author>
    <author>
      <name>Alexander C. Berg</name>
    </author>
    <arxiv:doi xmlns:arxiv="http://arxiv.org/schemas/atom">10.1007/978-3-319-46448-0_2</arxiv:doi>
    <link title="doi" href="http://dx.doi.org/10.1007/978-3-319-46448-0_2" rel="related"/>
    <arxiv:comment xmlns:arxiv="http://arxiv.org/schemas/atom">ECCV 2016</arxiv:comment>
    <link href="http://arxiv.org/abs/1234.5678v5" rel="alternate" type="text/html"/>
    <link title="pdf" href="http://arxiv.org/pdf/1234.5678v5" rel="related" type="application/pdf"/>
    <arxiv:primary_category xmlns:arxiv="http://arxiv.org/schemas/atom" term="cs.CV" scheme="http://arxiv.org/schemas/atom"/>
    <category term="cs.CV" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>
`

func TestArxivSpider_Import(t *testing.T) {
	var queryParams url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, arxivResponse)
		queryParams = r.URL.Query()
	}))
	defer ts.Close()

	arxivURL = ts.URL
	importer := ArxivSpider{}

	tts := map[string]struct {
		URL            string
		ExpectedParams map[string][]string
		Error          bool
	}{
		"abstract": {
			URL: "http://arxiv.org/abs/1234.5678v5",
			ExpectedParams: map[string][]string{
				"id_list":   []string{"1234.5678"},
				"sortBy":    []string{"submittedDate"},
				"sortOrder": []string{"descending"},
			},
			Error: false,
		},
		"pdf": {
			URL: "http://arxiv.org/pdf/1234.5678",
			ExpectedParams: map[string][]string{
				"id_list":   []string{"1234.5678"},
				"sortBy":    []string{"submittedDate"},
				"sortOrder": []string{"descending"},
			},
			Error: false,
		},
		"failing because not arxiv": {
			URL:            "http://medium.org/me/bookmarks",
			ExpectedParams: map[string][]string{},
			Error:          true,
		},
		"failing because could not extract id": {
			URL:            "http://arxiv.org/abs/not-an-id",
			ExpectedParams: map[string][]string{},
			Error:          true,
		},
	}
	for name, tt := range tts {
		queryParams = nil
		_, err := importer.Import(tt.URL)

		if err == nil && tt.Error {
			t.Errorf("%s - should have failed but did not", name)
		} else if err != nil && !tt.Error {
			t.Errorf("%s - should not have failed but did with error %v", name, err)
		} else if err != nil {
			// I don't know why but a wrapping deep equal does not seem to work...
			for k, v := range tt.ExpectedParams {
				if !reflect.DeepEqual(v, queryParams[k]) {
					t.Errorf("%s - invalid query params %s: expected %v got %v", name, k, v, queryParams[k])
				}
			}
			for k, v := range queryParams {
				if !reflect.DeepEqual(v, tt.ExpectedParams[k]) {
					t.Errorf("%s - invalid query params %s: expected %v got %v", name, k, tt.ExpectedParams[k], v)
				}
			}
		}
	}
}
