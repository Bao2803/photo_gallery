package controllers

import (
	"bao2803/photo_gallery/context"
	"bao2803/photo_gallery/models"
	"bao2803/photo_gallery/views"
	"fmt"
	"net/http"
)

func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		New:      views.NewView("bootstrap", "galleries/new"),
		ShowView: views.NewView("bootstrap", "galleries/show"),
		gs:       gs,
	}
}

type Galleries struct {
	New      *views.View
	ShowView *views.View
	gs       models.GalleryDB
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// Create POST /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}

	user := context.User(r.Context())
	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}
	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}
	fmt.Fprintln(w, gallery)
}
