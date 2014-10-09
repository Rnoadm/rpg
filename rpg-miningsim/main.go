package main

import (
	"bytes"
	"flag"
	"github.com/Rnoadm/rpg"
	"github.com/Rnoadm/rpg/gui"
	"github.com/Rnoadm/rpg/history"
	"github.com/Rnoadm/rpg/rpg-miningsim/res"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"time"
)

var (
	flagFilename = flag.String("f", "miningsim.sav", "filename for save file")
	flagReplay   = flag.Duration("replay", 0, "play back the game up to this point with this delay between frames")
)

func init() {
	rpg.MaxMessages = 1
}

func main() {
	flag.Parse()

	f, err := os.OpenFile(*flagFilename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	h := history.NewHistory(f)
	s, err := h.Seek(0, history.SeekEnd)
	if err == io.EOF {
		s = rpg.NewState()
		s.Atomic(func(s *rpg.State) bool {
			_, o := s.Create(MinedLocationsFactory)
			for x := int64(-2); x <= 2; x++ {
				for y := int64(-2); y <= 2; y++ {
					for z := int64(0); z <= 0; z++ {
						o.Component(MinedLocationsType).(*MinedLocations).Add(x, y, z)
					}
				}
			}
			_, o = s.Create(PlayerFactory, rpg.ContainerFactory, rpg.LocationFactory, rpg.MessagesFactory)
			_, pickaxe := s.Create(PickaxeFactory, rpg.LocationFactory)
			o.Component(rpg.ContainerType).(*rpg.Container).Add(pickaxe)
			return true
		})
		err = h.Append(s)
	}
	if err != nil {
		panic(err)
	}

	replayDone := make(chan struct{})
	if *flagReplay > 0 {
		h.Reset()
		go func(d time.Duration) {
			for {
				gui.Redraw()
				select {
				case <-replayDone:
					return
				default:
				}
				time.Sleep(d)
			}
		}(*flagReplay)
	}

	gui.Main("Mining Simulator 2014", &Handler{
		h:              h,
		s:              s,
		playerSprite:   LoadImage(res.PlayerPng),
		fontSprites:    LoadImage(res.FontPng),
		terrainSprites: LoadImage(res.TerrainPng),
		pickaxeCount:   LoadImage(res.PickcountPng),
		replayDone:     replayDone,
	})
}

func LoadImage(b []byte) *image.RGBA {
	src, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Rect, src, dst.Rect.Min, draw.Src)
	return dst
}
