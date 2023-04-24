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

var msPerTick uint32 = 15
var durationPerTick = int64(msPerTick) * int64(time.Millisecond)

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
	Paint()
}

type ImgDOB struct {
	BaseAnim
	ctx     *Ctx
	h       int32
	texture *sdl.Texture
	w       int32
	x       float64
	y       float64
}

func (i *ImgDOB) Tick(tick int64) {
	for j, anim := range i.Anims {
		if anim.Tick(tick) {
			i.Anims = append(i.Anims[:j], i.Anims[j+1:]...)
		}
	}
}

func (i *ImgDOB) Paint() {
	src := sdl.Rect{X: 0, Y: 0, W: i.w * 2, H: i.h * 2}
	dst := sdl.Rect{X: int32(i.x), Y: int32(i.y), W: i.w, H: i.h}
	i.ctx.Renderer.Copy(i.texture, &src, &dst)
}

type Anim interface {
	Tick(tick int64) bool
}

// BaseAnim has a DOB and can chain other anims
type BaseAnim struct {
	Dob   *ImgDOB
	Anims []Anim
}

func (b *BaseAnim) anim(a Anim) Anim {
	if len(b.Anims) == 0 {
		b.Anims = make([]Anim, 0)
	}
	b.Anims = append(b.Anims, a)
	return a
}

func (b *BaseAnim) Move(x float64, y float64, duration time.Duration) Anim {
	moveAnim := &MoveAnim{dob: b.Dob, endX: x, endY: y, duration: int64(duration)}
	return b.anim(moveAnim)
}

func (m *BaseAnim) Tick(tick int64) {
}

// Move Anim
type MoveAnim struct {
	BaseAnim
	deltaX    float64
	deltaY    float64
	dob       *ImgDOB
	duration  int64
	endX      float64
	endY      float64
	startTick int64
}

func (m *MoveAnim) Tick(tick int64) bool {
	if m.startTick == 0 {
		m.startTick = tick
		m.deltaX = m.endX - m.dob.x
		m.deltaY = m.endY - m.dob.y
	}
	pct := float64(tick-m.startTick) * float64(durationPerTick) / float64(m.duration)
	if pct > 1 {
		pct = 1
	}
	m.dob.x = m.endX - m.deltaX + pct*m.deltaX
	m.dob.y = m.endY - m.deltaY + pct*m.deltaY
	return pct == 1
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
		dob.BaseAnim.Dob = dob
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
	dobs []*ImgDOB
}

func (s *Stage) Tick(tick int64) {
	for i := 0; i < len(s.dobs); i++ {
		s.dobs[i].Tick(tick)
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

func (s *Stage) DOBsPut(d ...*ImgDOB) {
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
	s := &Stage{ctx: ctx, dobs: make([]*ImgDOB, 0)}

	// make the loader
	l := &EZLoader{ctx: ctx}
	l.imgs = make(map[string]*ImgDOB)
	l.snds = make(map[string]Snd)

	// add a block to the screen
	greenBlock, _ := l.ImgGet("img/block_green.png")
	s.DOBsPut(greenBlock)

	// animate it
	greenBlock.Move(120, 300, 3*time.Second)

	// loop until the user quits
	running := true
	var tick int64 = 1
	for running {
		s.Tick(tick)
		s.Paint()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
			}
		}
		sdl.Delay(msPerTick)
		tick++
	}
}
