package papernet

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"

	"github.com/bobinette/papernet/database"
	"github.com/bobinette/papernet/dot"
	"github.com/bobinette/papernet/models"
)

type WebHandler interface {
	Register(*gin.Engine)
}

type handler struct {
	db     database.DB
	search database.Search
}

func NewHandler(db database.DB, s database.Search) WebHandler {
	return &handler{
		db:     db,
		search: s,
	}
}

func (h *handler) Register(r *gin.Engine) {
	r.LoadHTMLGlob("./public/templates/*")

	r.Static("/public", "./public")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/papers")
	})

	r.GET("/papers", h.list)
	r.POST("/papers", h.new)

	r.GET("/papers/:id", h.show)
	r.POST("/papers/:id", h.save)

	r.GET("/papers/:id/edit", h.edit)
	r.GET("/papers/:id/graph", h.graph)
	r.POST("/papers/:id/delete", h.delete)
}

func (h *handler) show(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	ps, err := h.db.Get(id)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	} else if len(ps) == 0 {
		c.String(http.StatusNotFound, fmt.Sprintf("no Paper for id %d", id))
		return
	}
	p := ps[0]

	s := blackfriday.MarkdownCommon(p.Summary)
	var hp = struct {
		*models.Paper
		Summary template.HTML
	}{
		p,
		template.HTML(s),
	}
	c.HTML(http.StatusOK, "view.html", hp)
}

func (h *handler) list(c *gin.Context) {
	q := c.Query("q")
	var d = struct {
		Papers []*models.Paper
		Search string
	}{}
	if q != "" {
		ids, err := h.search.Find(q)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		ps, err := h.db.Get(ids...)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		d.Papers = ps
		d.Search = q
	} else {
		ps, err := h.db.List()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		d.Papers = ps
		d.Search = ""
	}

	c.HTML(http.StatusOK, "list.html", d)
}

func (h *handler) new(c *gin.Context) {
	var p models.Paper
	err := h.db.Insert(&p)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/papers/%d/edit", p.ID))
}

func (h *handler) edit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	ps, err := h.db.Get(id)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	} else if len(ps) == 0 {
		c.String(http.StatusNotFound, fmt.Sprintf("no Paper for id %d", id))
		return
	}
	p := ps[0]

	refs := strings.Join(func(a []int) []string {
		var s []string
		for _, i := range a {
			s = append(s, strconv.Itoa(i))
		}
		return s
	}(p.References), ",")
	var d = struct {
		*models.Paper
		References string
		Authors    string
		Tags       string
	}{
		p,
		refs,
		strings.Join(p.Authors, ","),
		strings.Join(p.Tags, ","),
	}

	c.HTML(http.StatusOK, "edit.html", d)
}

func (h *handler) save(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	var form struct {
		Title      string `form:"title" binding:"required"`
		Summary    string `form:"summary" binding:"required"`
		References string `form:"references" bindings:"required"`
		Authors    string `form:"authors" bindings:"required"`
		Tags       string `form:"tags" bindings:"required"`
		Read       string `form:"read" bindings:"required"`
	}
	err = c.Bind(&form)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("invalid form %v", err))
		return
	}

	refs, err := func(s string) ([]int, error) {
		if s == "" {
			return []int{}, nil
		}

		var ids []int
		for _, r := range strings.Split(s, ",") {
			// @TODO: better handle error
			id, err := strconv.Atoi(r)
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, nil
	}(form.References)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("invalid form %v", err))
		return
	}

	p := models.Paper{
		ID:         id,
		Title:      []byte(form.Title),
		Summary:    []byte(form.Summary),
		References: refs,
		Authors:    strings.Split(form.Authors, ","),
		Tags:       strings.Split(form.Tags, ","),
		Read:       form.Read == "on",
	}

	err = h.db.Update(&p)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = h.search.Index(&p)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/papers/%d", p.ID))
}

// deleteHandler is a lot of things but certainly not REST. Make it REST. Please.
func (h *handler) delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	err = h.db.Delete(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/papers")

}

func (h *handler) graph(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	ps, err := h.db.Get(id)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	} else if len(ps) == 0 {
		c.String(http.StatusNotFound, fmt.Sprintf("no Paper for id %d", id))
		return
	}
	p := ps[0]

	g := dot.NewGraph()
	node := p.Node()
	re := models.Edge{Style: models.EdgeReference}
	refs, err := h.db.Get(p.References...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	for _, rp := range refs {
		g.AddEdge(node, rp.Node(), re)
	}

	te := models.Edge{Style: models.EdgeTag}
	for i, t := range p.Tags {
		tn := models.Node{
			ID:    i,
			Label: t,
			Type:  models.NodeTag,
		}
		g.AddEdge(node, tn, te)
	}

	filedir := "./public/images"
	filename := fmt.Sprintf("%d_refs.dot", p.ID)
	filepath := path.Join(filedir, filename)
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer f.Close()

	_, err = g.WriteTo(f)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	svg, err := h.dotSVG(g)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var v = struct {
		Title string
		Graph template.HTML
	}{
		Title: string(p.Title),
		Graph: template.HTML(svg),
	}

	c.HTML(http.StatusOK, "graph.html", v)
}

func (h *handler) dotSVG(g *dot.Graph) (string, error) {
	f, err := ioutil.TempFile("./tmp", "papernet")
	if err != nil {
		return "", err
	}
	defer func() {
		f.Close()
		if err := os.Remove(f.Name()); err != nil {
			log.Printf("%v", err)
		}
	}()

	_, err = g.WriteTo(f)
	if err != nil {
		return "", err
	}

	dotCmd := exec.Command("dot", "-T", "svg", "-Kneato", f.Name())
	tailCmd := exec.Command("tail", "+4")
	r, w := io.Pipe()
	dotCmd.Stdout = w
	tailCmd.Stdin = r

	var b bytes.Buffer
	tailCmd.Stdout = &b

	err = dotCmd.Start()
	if err != nil {
		return "", nil
	}

	err = tailCmd.Start()
	if err != nil {
		return "", nil
	}

	err = dotCmd.Wait()
	if err != nil {
		return "", nil
	}

	err = w.Close()
	if err != nil {
		return "", nil
	}

	err = tailCmd.Wait()
	if err != nil {
		return "", nil
	}

	return b.String(), nil
}
