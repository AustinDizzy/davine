package page

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/hoisie/mustache"
)

//Page is the overall page content to render.
type Page struct {
	template string
	layout   string
	data     interface{}
}

const (
	titlePrefix        = "Davine - Open Data Analytics for Vine"
	defaultTemplateDir = "templates"
	defaultLayout      = "layout.html"
)

var (
	templateDir = path.Join(os.Getenv("PWD"), defaultTemplateDir)
)

//New returns a new page.
//If it is supplied no arguments, the page is blank and loaded with the
//default configured layout.
//If it is supplied a single string argument, it loads the page content
//with the supplied file found in the defaultTemplateDir directory.
//If it is supplied two strings, the first file is loaded as the layout
//and the second is loaded as the page content.
func New(tmpl ...string) *Page {
	p := new(Page)
	switch len(tmpl) {
	case 0:
		p.LoadLayout(defaultLayout)
	case 1:
		p.LoadTmpl(tmpl[0])
		p.LoadLayout(defaultLayout)
	case 2:
		p.LoadLayout(tmpl[0])
		p.LoadTmpl(tmpl[1])
	}
	return p
}

//LoadTmpl loads the supplied template, found in defaultTemplateDir, as
//the page content template.
func (p *Page) LoadTmpl(name string) {
	p.template = path.Join(templateDir, name)
}

//LoadLayout loads the supplied layout, found in defaultTemplateDir, as
//the page layout.
func (p *Page) LoadLayout(name string) {
	p.layout = path.Join(templateDir, name)
}

//LoadData loads the supplied data to be rendered as mustache variables
//in the page on Write.
func (p *Page) LoadData(data interface{}) {
	p.data = data
}

//Write writes the page contents to the supplied io.Writer.
func (p *Page) Write(w io.Writer) {
	//TODO: Alternate way to set page title.
	if p.data != nil {
		switch d := p.data.(type) {
		case map[string]interface{}:
			if d["title"] == nil || d["title"] == "" {
				d["title"] = titlePrefix
			} else {
				d["title"] = d["title"].(string) + " - " + titlePrefix
			}
		}
	}
	fmt.Fprint(w, mustache.RenderFileInLayout(p.template, p.layout, p.data))
}
