// radial: show data on a circle
// input is tab separated name,value pairs
// name is interpreted as a label, or with the -image flag, an image file.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/ajstarks/svgo/float"
)

const (
	fullcircle = math.Pi * 2
	topclock   = -(math.Pi / 2)
	smallest   = -math.MaxFloat64
	largest    = math.MaxFloat64
	stylefmt   = "font-size:%.1fpx;font-family:sans-serif;text-anchor:middle;fill:%s"
	cfmt       = "fill:%s;fill-opacity:%.2f"
	tfmt       = "font-size:200%;baseline-shift:-33%"
	lfmt       = "stroke-opacity:.25;stroke-width:1;stroke:"
)

type deps struct {
	name  string
	value float64
}

// vmap maps one range into another
func vmap(value, low1, high1, low2, high2 float64) float64 {
	return low2 + (high2-low2)*(value-low1)/(high1-low1)
}

// polar converts polar to Cartesian coordinates
func polar(cx, cy, r, t float64) (float64, float64) {
	return cx + r*math.Cos(t), cy + r*math.Sin(t)
}

// readata reads name, value pairs, and returns the min and max values
func readata(r io.Reader) ([]deps, float64, float64) {
	var data []deps
	var d deps
	max := smallest
	min := largest
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		t := scanner.Text()
		fields := strings.Split(t, "\t")
		if len(fields) < 2 {
			continue
		}
		d.name = fields[0]
		v, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			d.value = 0
		}

		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
		d.value = v
		data = append(data, d)
	}
	return data, min, max
}

// imagedim returns thw width and height of an image
func imagedim(fname string) (int, int) {
	f, err := os.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 0, 0
	}
	im, _, err := image.DecodeConfig(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", fname, err)
		return 0, 0
	}
	f.Close()
	return im.Width, im.Height
}

func main() {
	var width, height float64
	var cmin, cmax, rpct, bop, fs float64
	var bg, fg, color, title string
	var image bool

	flag.Float64Var(&width, "width", 1000, "canvas width")
	flag.Float64Var(&height, "height", 1000, "canvas height")
	flag.Float64Var(&cmin, "bmin", 7.0, "bubble min")
	flag.Float64Var(&cmax, "bmax", 70.0, "bubble max")
	flag.Float64Var(&rpct, "r", 30, "radius percentage")
	flag.Float64Var(&bop, "op", 40, "opacity percentage")
	flag.Float64Var(&fs, "fs", 10, "font size")
	flag.StringVar(&bg, "bg", "white", "background color")
	flag.StringVar(&fg, "fg", "black", "text color")
	flag.StringVar(&color, "color", "lightsteelblue", "bubble color")
	flag.StringVar(&title, "title", "", "title")
	flag.BoolVar(&image, "image", false, "names are image files")
	flag.Parse()

	canvas := svg.New(os.Stdout)
	data, dmin, dmax := readata(os.Stdin)

	midx, midy := width/2, height/2
	r := width * (rpct / 100)
	t := topclock
	step := fullcircle / float64(len(data))

	canvas.Start(width, height)
	canvas.Rect(0, 0, width, height, "fill:"+bg)
	canvas.Gstyle(fmt.Sprintf(stylefmt, fs, fg))

	for _, d := range data {
		cr := vmap(d.value, dmin, dmax, cmin, cmax)
		cx, cy := polar(midx, midy, r, t)
		tx, ty := polar(midx, midy, r+cmax+(fs*2.5), t)
		if image {
			w, h := imagedim(d.name)
			if w == 0 && h == 0 {
				continue
			}
			canvas.Image(cx-float64(w/2), cy-float64(h/2), w, h, d.name)
		} else {
			canvas.Circle(cx, cy, cr, fmt.Sprintf(cfmt, color, bop/100))
			canvas.Text(tx, ty, d.name)
			canvas.Text(cx, cy, fmt.Sprintf("%v", d.value), tfmt)
			canvas.Line(cx, cy, tx, ty, lfmt+fg)
		}
		t += step
	}

	if !image {
		canvas.Circle(midx, midy, r, "fill-opacity:.2;fill:"+color)
	}
	if len(title) > 0 {
		canvas.Text(midx, midy, title, "font-size:300%")
	}
	canvas.Gend()
	canvas.End()
}
