// Author: Jeremy Kassis (jkassis@gmail.com).
// Public domain software.
//
// A frogger game.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

// Snd wrapper for sounds
type Snd interface {
	Play() error
	Stop()
}

type WavSnd struct {
	Ctx     *Ctx
	Wav     *mix.Chunk
	playing bool
	channel int
}

func (w *WavSnd) Play() (err error) {
	if w.playing {
		return nil
	}
	channel, err := w.Wav.Play(0, -1)
	w.channel = channel
	return
}

func (w *WavSnd) Stop() {
	if !w.playing {
		return
	}
	mix.HaltChannel(w.channel)
}

// DOB Display Object renders to the screen and animates
type DOB interface {
	Tick()
	Paint()
}

type ImgDOB struct {
	ctx     *Ctx
	h       int32
	texture *sdl.Texture
	w       int32
	x       int32
	y       int32
}

func (i *ImgDOB) Tick() {

}

func (i *ImgDOB) Paint() {
	src := sdl.Rect{X: 0, Y: 0, W: i.w * 2, H: i.h * 2}
	dst := sdl.Rect{X: i.x, Y: i.y, W: i.w, H: i.h}
	i.ctx.Renderer.Copy(i.texture, &src, &dst)
}

// Loader loads assets
type Loader interface {
	DOB
	Get(spec string) (any, error)
	ImgGet(spec string) (*ImgDOB, error)
	SndGet(spec string) (*Snd, error)
}

// Ctx provides context for all DOBs
type Ctx struct {
	Loader   Loader
	Renderer *sdl.Renderer
	View     *sdl.Window
	ViewH    int32
	ViewW    int32
}

// EZLoader loads and caches assets
type EZLoader struct {
	ctx  *Ctx
	imgs map[string]*ImgDOB
	snds map[string]Snd
}

func (l *EZLoader) Get(spec string) (any, error) {
	parts := strings.Split(spec, ":")
	typ := parts[0]
	if typ == "img" {
		return l.ImgGet(parts[1])
	} else if typ == "snd" {
		return l.SndGet(parts[1])
	}

	return nil, fmt.Errorf("could not load type: %v", typ)
}

func (l *EZLoader) ImgGet(path string) (dob *ImgDOB, err error) {
	var ok bool
	dob, ok = l.imgs[path]
	if !ok {
		dob = &ImgDOB{ctx: l.ctx}
		dob.texture, err = img.LoadTexture(l.ctx.Renderer, path)
		if err != nil {
			fmt.Printf("could not load texture at %s: %v\n", path, err)
			return
		}
		_, _, dob.w, dob.h, err = dob.texture.Query()
		if err != nil {
			fmt.Printf("could not query texture at %s: %v\n", path, err)
			return nil, err
		}
	}
	return
}

func (l *EZLoader) SndGet(path string) (Snd, error) {
	snd, ok := l.snds[path]
	if !ok {
		wav, err := mix.LoadWAV(path)
		if err != nil {
			return nil, err
		}
		snd = &WavSnd{Ctx: l.ctx, Wav: wav}
	}
	return snd, nil
}

func CHECK(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// Stage is the root of the display tree
type Stage struct {
	ctx  *Ctx
	dobs []DOB
}

func (s *Stage) Tick() {
	for i := 0; i < len(s.dobs); i++ {
		s.dobs[i].Tick()
	}
}

func (s *Stage) Paint() {
	// clear the screen
	CHECK(s.ctx.Renderer.Clear())

	// render dobs
	for i := 0; i < len(s.dobs); i++ {
		s.dobs[i].Paint()
	}

	s.ctx.Renderer.Present()
}

func (s *Stage) DOBsPut(d ...DOB) {
	s.dobs = append(s.dobs, d...)
}

// main let's go
func main() {
	var err error
	rand.Seed(time.Now().UnixNano())

	// setup the ctx
	ctx := &Ctx{}
	ctx.ViewW = 800
	ctx.ViewH = 600

	// init sdl
	CHECK(sdl.Init(sdl.INIT_AUDIO | sdl.INIT_EVENTS | sdl.INIT_TIMER | sdl.INIT_VIDEO))
	defer sdl.Quit()
	CHECK(img.Init(img.INIT_PNG))
	mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT, mix.DEFAULT_CHANNELS, 4096)

	// make the window
	ctx.View, ctx.Renderer, err = sdl.CreateWindowAndRenderer(ctx.ViewW, ctx.ViewH, sdl.WINDOW_SHOWN)
	CHECK(err)
	defer ctx.View.Destroy()
	ctx.View.SetTitle("Frogger")

	// make the stage
	s := &Stage{ctx: ctx, dobs: make([]DOB, 0)}

	// make the loader
	l := &EZLoader{ctx: ctx}
	l.imgs = make(map[string]*ImgDOB)
	l.snds = make(map[string]Snd)

	// load a block and put it on screen
	greenBlock, _ := l.ImgGet("img/block_green.png")
	s.DOBsPut(greenBlock)

	// loop forever
	running := true
	for running {
		s.Tick()
		s.Paint()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
		sdl.Delay(16)
	}
}
