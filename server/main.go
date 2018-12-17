package main

import (
	"context"
	"github.com/mjibson/goon"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type (
	QuadkeyTile struct {
		maptile.Tile
	}
	Tanuki struct {
		ID        string `datastore:"-" goon:"id"`
		Name      string `datastore:",noindex"`
		Quadkey20 string
		Geo       appengine.GeoPoint `datastore:",noindex"`
	}
)

func main() {
	http.HandleFunc("/", Index)
	appengine.Main()
}

func Index(w http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)

	// get level
	u, err := url.Parse(req.RequestURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slv := u.Query().Get("lv")
	lv, err := strconv.Atoi(slv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b, err := toBound(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tiles := EstimateTiles(b, lv)
	entities, err := fetchTanukis(ctx, tiles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// creating a feature collection
	res, err := createGeoJson(b, tiles, entities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
	return
}

func createGeoJson(org orb.Bound, quadkeyTiles []QuadkeyTile, entities []*Tanuki) ([]byte, error) {
	fc := geojson.NewFeatureCollection()
	orgf := geojson.NewFeature(org.ToPolygon())
	orgf.Properties["name"] = "request range"
	fc = fc.Append(orgf)

	ts := maptile.Tiles{}
	ts.ToFeatureCollection()
	for _, t := range quadkeyTiles {
		tf := geojson.NewFeature(t.Bound().ToPolygon())
		tf.Properties["quadkey"] = t.QuadkeyString()
		fc = fc.Append(tf)
	}

	for _, e := range entities {
		p := orb.Point{e.Geo.Lng, e.Geo.Lng}
		ef := geojson.NewFeature(p)
		ef.Properties["name"] = e.Name
		fc = fc.Append(ef)
	}
	return fc.MarshalJSON()
}

func fetchTanukis(ctx context.Context, tiles []QuadkeyTile) ([]*Tanuki, error) {
	g := goon.FromContext(ctx)
	eg := errgroup.Group{}
	mu := new(sync.Mutex)
	res := []*Tanuki{}

	f := func(t QuadkeyTile) func() error {
		return func() error {
			qk1, qk2 := t.FixQuadkey()

			q := datastore.NewQuery(g.Kind(Tanuki{})).
				Filter("Quadkey20 >=", qk1).
				Filter("Quadkey20 <", qk2).
				KeysOnly()
			entities := []*Tanuki{}
			if _, err := g.GetAll(q, &entities); err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			res = append(res, entities...)
			return nil
		}
	}

	for _, t := range tiles {
		eg.Go(f(t))
	}
	err := eg.Wait()
	return res, err
}

func EstimateTiles(b orb.Bound, ilv int) []QuadkeyTile {
	lv := maptile.Zoom(ilv)
	minTile := maptile.At(b.Min, lv)
	maxTile := maptile.At(b.Max, lv)

	// 指定レベル内でtileがひとつだけ(指定レベルが範囲に対して比較的小さい)場合は
	// 範囲内のタイルを総なめする価値がないのですぐ返す
	if reflect.DeepEqual(minTile, maxTile) {
		return []QuadkeyTile{{Tile: minTile}}
	}

	res := []QuadkeyTile{}
	minX := float64(minTile.X)
	minY := float64(minTile.Y)
	maxX := float64(maxTile.X)
	maxY := float64(maxTile.Y)

	// 範囲内のタイルを総なめしてQuadkey文字列を取り出す
	for x := math.Min(minX, maxX); x <= math.Max(minX, maxX); x++ {
		for y := math.Min(minY, maxY); y <= math.Max(minY, maxY); y++ {
			tile := QuadkeyTile{Tile: maptile.New(uint32(x), uint32(y), lv)}
			// 四隅が指定範囲に含まれているかチェック
			if !tile.IsContained(b) {
				continue
			}
			res = append(res, tile)
		}
	}
	return res
}

func toBound(req *http.Request) (orb.Bound, error) {
	res := orb.Bound{}
	r := orb.Ring{}
	u, err := url.Parse(req.RequestURI)
	if err != nil {
		return res, err
	}

	for _, k := range []string{"lt", "rt", "rb", "lb"} {
		p, err := toPoint(u.Query().Get(k))
		if err != nil {
			return res, err
		}
		r = append(r, p)
	}
	r = append(r, r[0])
	return r.Bound(), nil
}
func toPoint(str string) (orb.Point, error) {
	res := orb.Point{}
	if str == "" {
		return res, errors.New("blank string")
	}

	sp := strings.Split(str, ",")
	if len(sp) != 2 {
		return res, errors.Errorf("invalid separate count: orig:%s", str)
	}
	flt := make([]float64, 2, 2)
	for i, v := range sp {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return res, errors.Wrapf(err, "cannot parse string to float: %s", v)
		}
		flt[i] = f
	}
	return orb.Point{flt[1], flt[0]}, nil
}

// quadkeyString returns quadkey string from tile.
func (t QuadkeyTile) QuadkeyString() string {
	// see the original logic; https://github.com/paulmach/go.geo/blob/master/point.go#L149
	s := strconv.FormatInt(int64(t.Quadkey()), 4)

	// for zero padding
	zeros := "000000000000000000000000000000"
	return zeros[:((int(t.Z)+1)-len(s))/2] + s
}
func (t QuadkeyTile) IsContained(b orb.Bound) bool {
	tb := t.Bound()
	for _, p := range []orb.Point{tb.LeftTop(), tb.RightBottom(), {tb.Left(), tb.Bottom()}, {tb.Right(), tb.Top()}} {
		if b.Contains(p) {
			return true
		}
	}
	return false
}
func (t QuadkeyTile) FixQuadkey() (string, string) {
	org := t.QuadkeyString()
	if 20 < len(org) {
		return "", ""
	}
	lv := len(org)
	// Quadkey Query Filter のMax値に利用するため、最後の数字に+1する。
	lastString := org[lv-1 : lv]
	lastInt, _ := strconv.Atoi(lastString)
	lastInt++
	res2 := org[0:lv-1] + strconv.Itoa(lastInt)

	diff := 20 - lv
	for i := 0; i < diff; i++ {
		org += "0"
		res2 += "0"
	}
	return org, res2
}
