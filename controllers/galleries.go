package controllers

import (
	"bao2803/photo_gallery/context"
	"bao2803/photo_gallery/models"
	"bao2803/photo_gallery/views"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	IndexGalleries = "index_galleries"
	ShowGallery    = "show_gallery"
	EditGallery    = "edit_gallery"
)

const (
	maxMultipartMem = 1 << 20 // 1 megabyte
)

func NewGalleries(gs models.GalleryService, is models.ImageService, r *mux.Router) *Galleries {
	return &Galleries{
		New:       views.NewView("bootstrap", "galleries/new"),
		ShowView:  views.NewView("bootstrap", "galleries/show"),
		EditView:  views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		gs:        gs,
		is:        is,
		r:         r,
	}
}

type Galleries struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        models.GalleryDB
	is        models.ImageService
	r         *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Galleries, error) {
	// Getting gallery's id
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handleUnknownError(err, w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	// Retrieving corresponding gallery from DB
	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			handleUnknownError(err, w, "Galleries not found", http.StatusNotFound)
		default:
			handleUnknownError(err, w, "Whoops! Something went wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}
	// Populated gallery with its images' paths
	images, _ := g.is.ByGalleryID(gallery.ID)
	gallery.Images = images
	return gallery, nil
}

func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		handleUnknownError(err, w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
}

// Show GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, r, vd)
}

// Create POST /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}

	user := context.User(r.Context())
	gallery := models.Galleries{
		Title:  form.Title,
		UserID: user.ID,
	}
	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}

	url, err := g.r.Get(EditGallery).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// Edit GET /galleries/:id/edit
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		handleUnknownError(err, w, "You do not have permission to edit this gallery", http.StatusForbidden)
		return
	}
	var vd views.Data
	vd.Yield = gallery
	g.EditView.Render(w, r, vd)
}

// Update POST /galleries/:id/update
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		handleUnknownError(err, w, "You do not have permission to edit this gallery", http.StatusForbidden)
		return
	}
	var vd views.Data
	vd.Yield = gallery
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	gallery.Title = form.Title
	err = g.gs.Update(gallery)
	if err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Galleries updated successfully",
		}
	}
	g.EditView.Render(w, r, vd)
}

// Delete POST /galleries/:id/delete
func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	// Lookup the gallery
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	// Verify user permission
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		handleUnknownError(err, w, "You do not have permission to edit this gallery", http.StatusForbidden)
		return
	}
	// Delete
	var vd views.Data
	err = g.gs.Delete(gallery.ID)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	url, err := g.r.Get(IndexGalleries).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// ImageUpload POST /galleries/:id/images
func (g *Galleries) ImageUpload(w http.ResponseWriter, r *http.Request) {
	// Lookup the gallery
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	// Verify user permission
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		handleUnknownError(err, w, "You do not have permission to edit this gallery", http.StatusForbidden)
		return
	}
	var vd views.Data
	vd.Yield = gallery
	// Parse Images
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	// Create directory to store images
	galleryPath := filepath.Join("images", "galleries", fmt.Sprintf("%v", gallery.ID))
	err = os.MkdirAll(galleryPath, 0755)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	// Process files
	files := r.MultipartForm.File["images"]
	for _, f := range files {
		func() { // Silence loop defer warning -> without this, ALL defer happens after ALL iteration
			// Open individual file
			file, err := f.Open()
			if err != nil {
				vd.SetAlert(err)
				g.EditView.Render(w, r, vd)
				return
			}
			defer file.Close()
			// Create file in filesystem
			err = g.is.Create(gallery.ID, file, f.Filename)
			if err != nil {
				vd.SetAlert(err)
				g.EditView.Render(w, r, vd)
				return
			}
		}()
	}
	//// Success, render alert
	//vd.Alert = &views.Alert{
	//	Level:   views.AlertLvlSuccess,
	//	Message: "Images successfully uploaded!",
	//}
	//g.EditView.Render(w, r, vd)
	// Success, redirect user to edit gallery -> this will update the image list
	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// ImageDelete POST /galleries/:id/images/:filename/delete
func (g *Galleries) ImageDelete(w http.ResponseWriter, r *http.Request) {
	// Lookup the gallery
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	// Verify user permission
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		handleUnknownError(err, w, "You do not have permission to edit this gallery", http.StatusForbidden)
		return
	}
	// Get filename from web request path
	filename := mux.Vars(r)["filename"]
	// Build the image model
	i := models.Image{
		GalleryID: gallery.ID,
		Filename:  filename,
	}
	// Delete the image
	err = g.is.Delete(&i)
	if err != nil {
		// Failure, rendering error widget
		var vd views.Data
		vd.Yield = gallery
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	// Success, redirect user to edit gallery page
	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		http.Redirect(w, r, "/galleries", http.StatusFound)
		log.Println(err)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}
