package world

import (
	"fmt"
	"image"
	"log"
	"math"
	"path/filepath"

	"github.com/hajimehoshi/ebiten"
	"github.com/peterhellberg/gfx"

	tiled "github.com/kyeett/go-tiled"
)

// Canvas holds sprite sources and images for rendering
type Canvas struct {
	sources   map[string]*ebiten.Image
	renderers map[string]*ebiten.Image
}

// NewCanvas returns a Canvas with initiated maps
func NewCanvas() Canvas {
	return Canvas{
		sources:   make(map[string]*ebiten.Image),
		renderers: make(map[string]*ebiten.Image),
	}
}

func (c *Canvas) LoadTileset(mp *tiled.Map, source, target string) {
	tileset := mp.Tilesets[0]
	for _, l := range mp.Layers {
		for i, t := range l.Tiles {
			if t.IsNil() {
				continue
			}

			sx, sy := i%mp.Width, i/mp.Width

			x, y := TilesheetCoords(tileset, t.ID)
			srcRect := image.Rect(0, 0, tileset.TileWidth, tileset.TileHeight).Add(image.Pt(x, y))

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-float64(tileset.TileWidth/2), -float64(tileset.TileHeight/2))
			if t.VerticalFlip {
				op.GeoM.Scale(1, -1)
			}
			if t.HorizontalFlip {
				op.GeoM.Scale(-1, 1)
			}
			if t.DiagonalFlip {
				op.GeoM.Rotate(3 * math.Pi / 2)
				op.GeoM.Scale(1, -1)
			}
			op.GeoM.Translate(float64(tileset.TileWidth/2), float64(tileset.TileHeight/2))
			op.GeoM.Translate(float64(sx*tileset.TileWidth), float64(sy*tileset.TileHeight))

			c.renderers[target].DrawImage(c.sources[source].SubImage(srcRect).(*ebiten.Image), op)
		}
	}
}

func TilesheetCoords(t *tiled.Tileset, ID uint32) (int, int) {
	y := (ID) / uint32(t.Columns)
	x := (ID) % uint32(t.Columns)
	return int(x) * t.TileWidth, int(y) * t.TileHeight
}

// LoadScene loads a Tiled map and tileset and saves the resulting images into a canvas, and objects into the ECS system
func (w *World) LoadScene(name string) error {
	filename := "assets/maps/trigger-map.tmx"
	dir := filepath.Dir(filename)

	// dir := filepath.Dir(filename)

	m, err := tiled.LoadFromFile(filename)
	if err != nil {
		return err
	}

	w.canvas.sources["sprite"] = loadImage(dir + "/" + m.Tilesets[0].Image.Source)
	img, err := ebiten.NewImage(m.Width*m.TileWidth, m.Height*m.TileHeight, ebiten.FilterDefault)
	if err != nil {
		return err
	}
	w.canvas.renderers["background"] = img

	w.canvas.LoadTileset(m, "sprite", "background")
	fmt.Println("load scene", img)
	return nil
}

func loadImage(filename string) *ebiten.Image {
	// tilesetImg, err := gfx.DecodePNG(assets.FileReaderFatal(path))
	img, err := gfx.OpenPNG(filename)
	if err != nil {
		log.Fatal(err)
	}
	ebitenImg, err := ebiten.NewImageFromImage(img, ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	return ebitenImg
}