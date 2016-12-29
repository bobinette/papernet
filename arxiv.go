package papernet

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var arxivSummaryPipe = CleaningPipe(
	strings.TrimSpace,
	OneLine,
	strings.TrimSpace,
)

var arxivCategories = map[string]string{
	"stat.AP":            "Statistics - Applications",
	"stat.CO":            "Statistics - Computation",
	"stat.ML":            "Statistics - Machine Learning",
	"stat.ME":            "Statistics - Methodology",
	"stat.TH":            "Statistics - Theory",
	"q-bio.BM":           "Quantitative Biology - Biomolecules",
	"q-bio.CB":           "Quantitative Biology - Cell Behavior",
	"q-bio.GN":           "Quantitative Biology - Genomics",
	"q-bio.MN":           "Quantitative Biology - Molecular Networks",
	"q-bio.NC":           "Quantitative Biology - Neurons and Cognition",
	"q-bio.OT":           "Quantitative Biology - Other",
	"q-bio.PE":           "Quantitative Biology - Populations and Evolution",
	"q-bio.QM":           "Quantitative Biology - Quantitative Methods",
	"q-bio.SC":           "Quantitative Biology - Subcellular Processes",
	"q-bio.TO":           "Quantitative Biology - Tissues and Organs",
	"cs.AR":              "Computer Science - Architecture",
	"cs.AI":              "Computer Science - Artificial Intelligence",
	"cs.CL":              "Computer Science - Computation and Language",
	"cs.CC":              "Computer Science - Computational Complexity",
	"cs.CE":              "Computer Science - Computational Engineering; Finance; and Science",
	"cs.CG":              "Computer Science - Computational Geometry",
	"cs.GT":              "Computer Science - Computer Science and Game Theory",
	"cs.CV":              "Computer Science - Computer Vision and Pattern Recognition",
	"cs.CY":              "Computer Science - Computers and Society",
	"cs.CR":              "Computer Science - Cryptography and Security",
	"cs.DS":              "Computer Science - Data Structures and Algorithms",
	"cs.DB":              "Computer Science - Databases",
	"cs.DL":              "Computer Science - Digital Libraries",
	"cs.DM":              "Computer Science - Discrete Mathematics",
	"cs.DC":              "Computer Science - Distributed; Parallel; and Cluster Computing",
	"cs.GL":              "Computer Science - General Literature",
	"cs.GR":              "Computer Science - Graphics",
	"cs.HC":              "Computer Science - Human-Computer Interaction",
	"cs.IR":              "Computer Science - Information Retrieval",
	"cs.IT":              "Computer Science - Information Theory",
	"cs.LG":              "Computer Science - Learning",
	"cs.LO":              "Computer Science - Logic in Computer Science",
	"cs.MS":              "Computer Science - Mathematical Software",
	"cs.MA":              "Computer Science - Multiagent Systems",
	"cs.MM":              "Computer Science - Multimedia",
	"cs.NI":              "Computer Science - Networking and Internet Architecture",
	"cs.NE":              "Computer Science - Neural and Evolutionary Computing",
	"cs.NA":              "Computer Science - Numerical Analysis",
	"cs.OS":              "Computer Science - Operating Systems",
	"cs.OH":              "Computer Science - Other",
	"cs.PF":              "Computer Science - Performance",
	"cs.PL":              "Computer Science - Programming Languages",
	"cs.RO":              "Computer Science - Robotics",
	"cs.SE":              "Computer Science - Software Engineering",
	"cs.SD":              "Computer Science - Sound",
	"cs.SC":              "Computer Science - Symbolic Computation",
	"nlin.AO":            "Nonlinear Sciences - Adaptation and Self-Organizing Systems",
	"nlin.CG":            "Nonlinear Sciences - Cellular Automata and Lattice Gases",
	"nlin.CD":            "Nonlinear Sciences - Chaotic Dynamics",
	"nlin.SI":            "Nonlinear Sciences - Exactly Solvable and Integrable Systems",
	"nlin.PS":            "Nonlinear Sciences - Pattern Formation and Solitons",
	"math.AG":            "Mathematics - Algebraic Geometry",
	"math.AT":            "Mathematics - Algebraic Topology",
	"math.AP":            "Mathematics - Analysis of PDEs",
	"math.CT":            "Mathematics - Category Theory",
	"math.CA":            "Mathematics - Classical Analysis and ODEs",
	"math.CO":            "Mathematics - Combinatorics",
	"math.AC":            "Mathematics - Commutative Algebra",
	"math.CV":            "Mathematics - Complex Variables",
	"math.DG":            "Mathematics - Differential Geometry",
	"math.DS":            "Mathematics - Dynamical Systems",
	"math.FA":            "Mathematics - Functional Analysis",
	"math.GM":            "Mathematics - General Mathematics",
	"math.GN":            "Mathematics - General Topology",
	"math.GT":            "Mathematics - Geometric Topology",
	"math.GR":            "Mathematics - Group Theory",
	"math.HO":            "Mathematics - History and Overview",
	"math.IT":            "Mathematics - Information Theory",
	"math.KT":            "Mathematics - K-Theory and Homology",
	"math.LO":            "Mathematics - Logic",
	"math.MP":            "Mathematics - Mathematical Physics",
	"math.MG":            "Mathematics - Metric Geometry",
	"math.NT":            "Mathematics - Number Theory",
	"math.NA":            "Mathematics - Numerical Analysis",
	"math.OA":            "Mathematics - Operator Algebras",
	"math.OC":            "Mathematics - Optimization and Control",
	"math.PR":            "Mathematics - Probability",
	"math.QA":            "Mathematics - Quantum Algebra",
	"math.RT":            "Mathematics - Representation Theory",
	"math.RA":            "Mathematics - Rings and Algebras",
	"math.SP":            "Mathematics - Spectral Theory",
	"math.ST":            "Mathematics - Statistics",
	"math.SG":            "Mathematics - Symplectic Geometry",
	"astro-ph":           "Astrophysics",
	"cond-mat.dis-nn":    "Physics - Disordered Systems and Neural Networks",
	"cond-mat.mes-hall":  "Physics - Mesoscopic Systems and Quantum Hall Effect",
	"cond-mat.mtrl-sci":  "Physics - Materials Science",
	"cond-mat.other":     "Physics - Other",
	"cond-mat.soft":      "Physics - Soft Condensed Matter",
	"cond-mat.stat-mech": "Physics - Statistical Mechanics",
	"cond-mat.str-el":    "Physics - Strongly Correlated Electrons",
	"cond-mat.supr-con":  "Physics - Superconductivity",
	"gr-qc":              "General Relativity and Quantum Cosmology",
	"hep-ex":             "High Energy Physics - Experiment",
	"hep-lat":            "High Energy Physics - Lattice",
	"hep-ph":             "High Energy Physics - Phenomenology",
	"hep-th":             "High Energy Physics - Theory",
	"math-ph":            "Mathematical Physics",
	"nucl-ex":            "Nuclear Experiment",
	"nucl-th":            "Nuclear Theory",
	"physics.acc-ph":     "Physics - Accelerator Physics",
	"physics.ao-ph":      "Physics - Atmospheric and Oceanic Physics",
	"physics.atom-ph":    "Physics - Atomic Physics",
	"physics.atm-clus":   "Physics - Atomic and Molecular Clusters",
	"physics.bio-ph":     "Physics - Biological Physics",
	"physics.chem-ph":    "Physics - Chemical Physics",
	"physics.class-ph":   "Physics - Classical Physics",
	"physics.comp-ph":    "Physics - Computational Physics",
	"physics.data-an":    "Physics - Data Analysis; Statistics and Probability",
	"physics.flu-dyn":    "Physics - Fluid Dynamics",
	"physics.gen-ph":     "Physics - General Physics",
	"physics.geo-ph":     "Physics - Geophysics",
	"physics.hist-ph":    "Physics - History of Physics",
	"physics.ins-det":    "Physics - Instrumentation and Detectors",
	"physics.med-ph":     "Physics - Medical Physics",
	"physics.optics":     "Physics - Optics",
	"physics.ed-ph":      "Physics - Physics Education",
	"physics.soc-ph":     "Physics - Physics and Society",
	"physics.plasm-ph":   "Physics - Plasma Physics",
	"physics.pop-ph":     "Physics - Popular Physics",
	"physics.space-ph":   "Physics - Space Physics",
	"quant-ph":           "Quantum Physics",
}

type ArxivSearch struct {
	Q          string
	Start      int
	MaxResults int
}

type ArxivResult struct {
	Papers     []*Paper
	Pagination Pagination
}

type ArxivSpider struct {
	Client *http.Client
}

func (s *ArxivSpider) Search(search ArxivSearch) (ArxivResult, error) {
	u, _ := url.Parse("http://export.arxiv.org/api/query")

	query := u.Query()

	if search.Q != "" {
		query.Add("search_query", fmt.Sprintf("all:%s", search.Q))
	}
	if search.Start > 0 {
		query.Add("start", strconv.Itoa(search.Start))
	}
	if search.MaxResults > 0 {
		query.Add("max_results", strconv.Itoa(search.MaxResults))
	}

	query.Add("sortBy", "submittedDate")
	query.Add("sortOrder", "descending")

	u.RawQuery = query.Encode()

	resp, err := s.Client.Get(u.String())
	if err != nil {
		return ArxivResult{}, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ArxivResult{}, err
	}

	r := struct {
		Title string `xml:"title"`
		ID    string `xml:"id"`
		Total struct {
			Value uint64 `xml:",chardata"`
		} `xml:"totalResults"`
		Offset struct {
			Value uint64 `xml:",chardata"`
		} `xml:"startIndex"`
		Limit struct {
			Value uint64 `xml:",chardata"`
		} `xml:"itemsPerPage"`
		Entries []struct {
			Title   string `xml:"title"`
			ID      string `xml:"id"`
			Summary string `xml:"summary"`
			Links   []struct {
				HRef string `xml:"href,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
			Categories []struct {
				Term string `xml:"term,attr"`
			} `xml:"category"`
			Published time.Time `xml:"published"`
			Updated   time.Time `xml:"updated"`
		} `xml:"entry"`
	}{}
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return ArxivResult{}, err
	}

	papers := make([]*Paper, len(r.Entries))
	for i, entry := range r.Entries {
		tags := make([]string, 0, len(entry.Categories))
		for _, cat := range entry.Categories {
			tag, ok := arxivCategories[cat.Term]
			if ok {
				tags = append(tags, tag)
			}
		}

		papers[i] = &Paper{
			Title:   entry.Title,
			Summary: arxivSummaryPipe(entry.Summary),
			References: []string{
				entry.Links[0].HRef, // link to arXiv
				entry.Links[1].HRef, // PDF
			},
			Tags:      tags,
			CreatedAt: entry.Published,
			UpdatedAt: entry.Updated,
			ArxivID:   entry.ID,
		}
	}

	return ArxivResult{
		Papers: papers,
		Pagination: Pagination{
			Total:  r.Total.Value,
			Limit:  r.Limit.Value,
			Offset: r.Offset.Value,
		},
	}, nil
}
