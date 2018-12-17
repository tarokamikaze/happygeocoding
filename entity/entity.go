package entity

import "google.golang.org/appengine"

type(
	Tanuki struct {
		ID        string `datastore:"-" goon:"id"`
		Name      string `datastore:",noindex"  validate:"required"`
		Quadkey20 string
		Geo       appengine.GeoPoint `datastore:",noindex" validate:"required"`
	}
)
