package main

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"strconv"
)

// quadkeyString returns quadkey string from tile.
func quadkeyString(t maptile.Tile) string {
	// see the original logic; https://github.com/paulmach/go.geo/blob/master/point.go#L149
	s := strconv.FormatInt(int64(t.Quadkey()), 4)

	// for zero padding
	zeros := "000000000000000000000000000000"
	return zeros[:((int(t.Z)+1)-len(s))/2] + s
}

func main() {
	name := "皇居"
	address := "東京都千代田区千代田１−１"

	// Creating a point
	lat := 35.685323
	lng := 139.752768
	p := orb.Point{lng, lat}

	// Change to a maptile.
	t := maptile.At(p, 17)
	println(quadkeyString(t)) // 13300211231022032

	// viewing as geojson
	c := geojson.NewFeatureCollection()
	gp := geojson.NewFeature(p)
	gp.Properties["name"] = name
	gp.Properties["address"] = address

	gt := geojson.NewFeature(t.Bound())
	gt.Properties["quadkey"] = quadkeyString(t)

	c = c.Append(gp).Append(gt)
	b, _ := c.MarshalJSON()
	println(string(b)) //{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[139.752768,35.685323]},"properties":{"address":"東京都千代田区千代田１−１","name":"皇居"}},{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[139.7515869140625,35.684071533140965],[139.75433349609375,35.684071533140965],[139.75433349609375,35.68630240145626],[139.7515869140625,35.68630240145626],[139.7515869140625,35.684071533140965]]]},"properties":{"quadkey":"13300211231022032"}}]}
}

//	// Creating a tile by a point
//	t := maptile.At(p, 16)
//	b, _ = geojson.NewFeature(t.Bound()).MarshalJSON()
//	//fmt.Printf("quadkey16: %s geojson: %s\n", quadkeyString(t), string(b))
//
//	ms := MapScreen{
//		Tl: GeoPoint{
//			Latitude:  35.6949464269382,
//			Longitude: 139.7493553161621,
//		},
//		Tr: GeoPoint{
//			Latitude:  35.680655395139816,
//			Longitude: 139.73888397216797,
//		},
//		Br: GeoPoint{
//			Latitude:  35.672358451082985,
//			Longitude: 139.7556209564209,
//		},
//		Bl: GeoPoint{
//			Latitude:  35.68546583314841,
//			Longitude: 139.76652145385742,
//		},
//	}
//	quadkeys := EstimateQuadkey(ms, 17)
//	println(strings.Join(quadkeys, ","))
//}
//type (
//	GeoPoint struct {
//		Latitude  float64
//		Longitude float64
//	}
//	MapScreen struct {
//		Tl GeoPoint // top left
//		Tr GeoPoint // top right
//		Bl GeoPoint // bottom left
//		Br GeoPoint // bottom right
//	}
//)
//// ToPoint converts GeoPoint to orb.Point
//func (p GeoPoint) ToPoint() orb.Point {
//	return orb.Point{p.Longitude, p.Latitude}
//}
//
//// EstimateQuadkey returns quadkeys contained in the map screen.
//func EstimateQuadkey(arg MapScreen, ilv int) []string {
//
//	r := orb.Ring{arg.Bl.ToPoint()}
//	for _, gp := range []GeoPoint{arg.Tl, arg.Tr, arg.Br, arg.Bl} {
//		r = append(r, gp.ToPoint())
//	}
//
//	b := r.Bound()
//
//
//	lv := maptile.Zoom(ilv)
//	minTile := maptile.At(b.Min, lv)
//	maxTile := maptile.At(b.Max, lv)
//
//	// 指定レベル内でtileがひとつだけ(指定レベルが範囲に対して比較的小さい)場合は
//	// 範囲内のタイルを総なめする価値がない
//	if reflect.DeepEqual(minTile, maxTile) {
//		println("only one quadkey")
//		return []string{quadkeyString(minTile)}
//	}
//
//	minX := float64(minTile.X)
//	minY := float64(minTile.Y)
//	maxX := float64(maxTile.X)
//	maxY := float64(maxTile.Y)
//
//	println(planar.RingContains(r,minTile.Bound().Min))
//
//	tiles := maptile.Tiles{}
//	res := []string{}
//	fmt.Printf("min %v:%v max %v:%v\n", minTile.X, minTile.Y, maxTile.X, maxTile.Y)
//	// 範囲内のタイルを総なめしてQuadkey文字列を取り出す
//	for x := math.Min(minX, maxX); x <= math.Max(minX, maxX); x++ {
//		for y := math.Min(minY, maxY); y <= math.Max(minY, maxY); y++ {
//			tile := maptile.New(uint32(x), uint32(y), lv)
//			res = append(res, quadkeyString(tile))
//			tiles = append(tiles, tile)
//		}
//	}
//	println("map tile is ", len(tiles))
//	col := tiles.ToFeatureCollection()
//	//col := geojson.NewFeatureCollection()
//	//col.Append(geojson.NewFeature(b))
//	col.Append(geojson.NewFeature(r))
//	bt, _ := col.MarshalJSON()
//	println(string(bt))
//
//	return res
//}
//
