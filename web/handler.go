package papernet

import (
	"fmt"
	"html/template"
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
	db database.DB
}

func NewHandler(db database.DB) WebHandler {
	return &handler{
		db: db,
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
	r.GET("/papers/:id/refgraph", h.referencesGraph)
	r.POST("/papers/:id/delete", h.delete)
}

func (h *handler) show(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	p, err := h.db.Get(id)
	if err != nil {
		log.Printf("retrieve paper: %v", err)
		return
	}

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
	ps, err := h.db.List()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "list.html", ps)
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

	p, err := h.db.Get(id)
	if err != nil {
		log.Printf("retrieve paper: %v", err)
		return
	}

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
	}{
		p,
		refs,
		strings.Join(p.Authors, ","),
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
		Read       string `form:"read" bindings:"required"`
	}
	err = c.Bind(&form)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("invalid form %v", err))
		return
	}
	log.Println(form)

	refs, err := func(s string) ([]int, error) {
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
		Read:       form.Read == "on",
	}

	err = h.db.Update(&p)
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

func (h *handler) referencesGraph(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	p, err := h.db.Get(id)
	if err != nil {
		log.Printf("retrieve paper: %v", err)
		return
	}

	filedir := "./public/images"
	filename := fmt.Sprintf("%d_refs.dot", p.ID)
	filepath := path.Join(filedir, filename)
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	g := dot.NewGraph()
	for _, r := range p.References {
		g.AddEdge(p.ID, r)
	}
	g.WriteTo(f)

	err = exec.Command("dot", "-T", "svg", "-O", filepath).Run()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var v = struct {
		Title     string
		GraphPath string
	}{
		Title:     string(p.Title),
		GraphPath: fmt.Sprintf("%s.svg", filename),
	}

	c.HTML(http.StatusOK, "refgraph.html", v)
}
