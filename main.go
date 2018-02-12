package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	nykakin_quantize "github.com/Nykakin/quantize"
	"github.com/RobCherry/vibrant"
	joshdk_quantize "github.com/joshdk/quantize"
	"github.com/marekm4/color-extractor"
	"github.com/nfnt/resize"
)

const (
	HTML = `
		<!DOCTYPE html>
		<html>
		<head>
		<style>
		img{
		    width:100%%;
		    max-width:200px;
		}
		svg {
		    border: 2px solid black;
		    border-radius: 5px;
		}
		</style>
		</head>
		<body>
		<table style="width:100%%">
		<tr>
		    <th></th>
	    	<th>github.com/Nykakin/quantize</th>
			<th>github.com/marekm4/color-extractor</th>
	    	<th>github.com/RobCherry/vibrant</th>
			<th>github.com/joshdk/quantize</th>
		</tr>
		%s
		</table>
		</body>
		</html>
	`

	ROW = `
		<tr>
		    <td>%s<br/>%s</td>
	    	<td>%s</td>
	    	<td>%s</td>
	    	<td>%s</td>
	    	<td>%s</td>
		</tr>
	`

	SVG = `
		<svg width="250" height="50">
		%s
		</svg>
	`
	RECT = "<rect x=\"%d\" width=\"50\" height=\"50\" style=\"fill:rgb(%d,%d,%d)\" />"
)

func nykakin(img image.Image) string {
	colors, err := nykakin_quantize.DominantColors(img, 5)
	if err != nil {
		panic(err)
	}

	rects := []string{}
	for index, clr := range colors {
		rects = append(rects, fmt.Sprintf(RECT, index*50, clr.R, clr.G, clr.B))
	}

	return fmt.Sprintf(SVG, strings.Join(rects, "\n"))
}

func marekm4(img image.Image) string {
	colors := color_extractor.ExtractColors(img)

	rects := []string{}
	for index, color := range colors {
		r, g, b, _ := color.RGBA()

		rects = append(rects, fmt.Sprintf(RECT, index*50, r>>8, g>>8, b>>8))
	}

	return fmt.Sprintf(SVG, strings.Join(rects, "\n"))
}

func joshdk(img image.Image) string {
	colors := joshdk_quantize.Image(img, 5)

	rects := []string{}
	for index, clr := range colors {
		rects = append(rects, fmt.Sprintf(RECT, index*50, clr.R, clr.G, clr.B))
	}

	return fmt.Sprintf(SVG, strings.Join(rects, "\n"))
}

func robCherry(img image.Image) string {
	palette := vibrant.NewPaletteBuilder(img).MaximumColorCount(uint32(5)).Generate()
	swatches := palette.Swatches()
	sort.Sort(populationSwatchSorter(swatches))

	rects := []string{}
	for index, swatch := range swatches {

		clr := color.NRGBAModel.Convert(swatch.Color()).(color.NRGBA)
		rects = append(rects, fmt.Sprintf(RECT, index*50, clr.R, clr.G, clr.B))
	}

	return fmt.Sprintf(SVG, strings.Join(rects, "\n"))
}

func process(img image.Image, filename string) string {
	return fmt.Sprintf(
		ROW,
		filename,
		fmt.Sprintf("<img src=\"%s\">", filename),
		nykakin(img),
		marekm4(img),
		robCherry(img),
		joshdk(img),
	)
}

func main() {
	files, err := ioutil.ReadDir("./images/")
	if err != nil {
		panic(err)
	}

	rows := []string{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".go") || strings.HasSuffix(f.Name(), ".html") {
			continue
		}

		f, err := os.Open("images/" + f.Name())
		if err != nil {
			panic(err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			panic(err)
		}

		img = resize.Resize(100, 100, img, resize.NearestNeighbor)
		rows = append(rows, process(img, f.Name()))
	}

	fo, err := os.Create("index.html")
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	_, err = io.Copy(fo, strings.NewReader(fmt.Sprintf(HTML, strings.Join(rows, ""))))
	if err != nil {
		panic(err)
	}
}

type populationSwatchSorter []*vibrant.Swatch

func (p populationSwatchSorter) Len() int           { return len(p) }
func (p populationSwatchSorter) Less(i, j int) bool { return p[i].Population() > p[j].Population() }
func (p populationSwatchSorter) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }