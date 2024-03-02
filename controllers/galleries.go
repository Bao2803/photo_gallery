package controllers

import (
	"net/http"

	"bao2803/photo_gallery/views"
)

func NewGalleries() *Galleries {
	return &Galleries{
		Gallery: views.NewView("bootstrap", "galleries/new"),
	}
}

type Galleries struct {
	Gallery *views.View
}

func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	err := g.Gallery.Render(w, nil)
	if err != nil {
		panic(err)
	}
}
