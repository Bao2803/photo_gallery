package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Image is used to represent images stored in a Gallery.
// Image is NOT stored in the database, and instead references data stored on disk.
type Image struct {
	GalleryID uint
	Filename  string
}

// RelativePath is used to build the path to this image on our local disk,
// relative to where the Go application is run from
func (i *Image) RelativePath() string {
	galleryID := fmt.Sprintf("%v", i.GalleryID)
	return filepath.ToSlash(filepath.Join("images", "galleries", galleryID, i.Filename))
}

// Path is used to build the absolute path used to reference this image via web request.
func (i *Image) Path() string {
	return "/" + i.RelativePath()
}

type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(i *Image) error
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct{}

func (is *imageService) imageDir(galleryID uint) string {
	return filepath.Join("images", "galleries", fmt.Sprintf("%v", galleryID))
}

func (is *imageService) mkImageDir(galleryID uint) (string, error) {
	galleryPath := is.imageDir(galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}

func (is *imageService) Create(galleryID uint, r io.Reader, filename string) error {
	path, err := is.mkImageDir(galleryID)
	if err != nil {
		return err
	}
	// Create a destination file
	dst, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}
	defer dst.Close()
	// Copy uploaded file data to the destination file
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.imageDir(galleryID)
	strings, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}
	ret := make([]Image, len(strings))
	for i, imgStr := range strings {
		ret[i] = Image{
			GalleryID: galleryID,
			Filename:  filepath.Base(imgStr),
		}
	}
	return ret, nil
}

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}
