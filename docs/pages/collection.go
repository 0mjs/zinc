package pages

import "html/template"

type Page struct {
	Title    string
	Subtitle string
	Content  template.HTML
}

func (p *Page) Find(pages map[string]Page, name string) Page {
	introPage := pages["introduction"]
	if page, exists := pages[name]; exists {
		return page
	}
	return introPage
}
