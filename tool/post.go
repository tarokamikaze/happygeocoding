package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/pkg/errors"
	"github.com/tarokamikaze/happygeocoding/entity"
	"golang.org/x/sync/errgroup"
	"google.golang.org/appengine"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var names = []string{
	"正吉",
	"おキヨ",
	"鶴亀和尚",
	"おろく婆",
	"権太",
	"青左衛門	",
	"ぽん吉",
	"文太",
	"玉三郎",
	"佐助",
	"六代目金長",
	"太三朗禿狸",
	"隠神刑部",
	"お玉",
	"小春",
	"花子",
	"お福",
}

// go run tool/post.go {entity_num} {host(default:http://localhost:8081)}
func main() {
	flag.Parse()

	max, _ := strconv.Atoi(flag.Arg(0))
	nameNum := len(names) - 1
	if max > nameNum {
		max = nameNum
	}
	host := flag.Arg(1)
	if host == "" {
		host = "http://localhost:8081"
	}

	sw := orb.Point{139.733427, 35.674296}
	ne := orb.Point{139.764499, 35.694091}

	xDiff := ne.X() - sw.X()
	yDiff := ne.Y() - sw.Y()
	rand.Seed(time.Now().UnixNano())

	eg := errgroup.Group{}

	u := fmt.Sprintf("%s/tanuki", host)
	for i := 0; i < max; i++ {
		name := names[i]
		p := orb.Point{
			rand.Float64()*xDiff + sw.X(),
			rand.Float64()*yDiff + sw.Y(),
		}

		e := &entity.Tanuki{
			Name: name,
			Geo: appengine.GeoPoint{
				Lat: p.Lat(),
				Lng: p.Lon(),
			},
		}
		eg.Go(post(e, u))
	}
	if err := eg.Wait(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("succeess to post %d tanukis\n", max)
}
func post(e *entity.Tanuki, u string) func() error {
	return func() error {
		j, err := json.Marshal(e)
		if err != nil {
			return err
		}
		res, err := http.Post(u, "application/json", bytes.NewReader(j))
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			bd, _ := ioutil.ReadAll(res.Body)
			println(string(bd))
			return errors.Errorf("invalid response code: %d error:%s", res.StatusCode, string(bd))
		}
		return err
	}
}
