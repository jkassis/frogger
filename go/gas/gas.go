// gas is the game animation system
package gas

import (
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

func Init() (err error) {
	// init sdl
	err = sdl.Init(sdl.INIT_AUDIO | sdl.INIT_EVENTS | sdl.INIT_TIMER | sdl.INIT_VIDEO)
	if err != nil {
		return err
	}
	err = img.Init(img.INIT_PNG)
	if err != nil {
		return err
	}
	return mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT, mix.DEFAULT_CHANNELS, 4096)
}

func Destroy() {
	sdl.Quit()
}

func MakeView(w, h int32, title string) (view *View, err error) {
	view = &View{W: w, H: h, Title: "Frogger"}
	return view, view.Init()
}

// Tex wrapper
type Tex struct {
	SDLTexture *sdl.Texture
	H          int32
	W          int32
}

// Wav wrapper for sounds
type Wav struct {
	View    *View
	Wav     *mix.Chunk
	playing bool
	channel int
}

func (w *Wav) Play() (err error) {
	if w.playing {
		return nil
	}
	channel, err := w.Wav.Play(0, -1)
	w.channel = channel
	return
}

func (w *Wav) Stop() {
	if !w.playing {
		return
	}
	mix.HaltChannel(w.channel)
}

// Dob Display Object renders to the screen and animates
type Dob struct {
	BaseAn
	h       int32
	id      int32
	spawn   map[int32]*Dob
	stage   *Stage
	texture *Tex
	w       int32
	x       float32
	y       float32
	zoom    float32
}

func (d *Dob) Tick(tick int32) {
	for j, anim := range d.anQ {
		if anim.Tick(tick) {
			d.anQ = append(d.anQ[:j], d.anQ[j+1:]...)
			d.anQ = append(d.anQ, anim.AnQ()...)
		}
	}
}

func (d *Dob) Paint() {
	src := sdl.Rect{X: 0, Y: 0, W: d.w, H: d.h}
	dst := sdl.Rect{X: int32(d.x - d.zoom*float32(d.w)/2), Y: int32(d.y - d.zoom*float32(d.h)/2), W: int32(d.zoom * float32(d.w)), H: int32(d.zoom * float32(d.h))}
	d.stage.view.Renderer.Copy(d.texture.SDLTexture, &src, &dst)
}

func (d *Dob) Spawn(path string) (dob *Dob, err error) {
	dob = &Dob{
		id:    d.stage.SpawnId,
		stage: d.stage,
		zoom:  1,
	}
	d.stage.SpawnId++
	dob.texture, _ = d.stage.view.TextureLoad(path)
	dob.h = dob.texture.H
	dob.w = dob.texture.W
	dob.BaseAn.Dob = dob
	if d.spawn == nil {
		d.spawn = make(map[int32]*Dob, 0)
	}

	d.spawn[dob.id] = dob
	return
}

// An animates
type An interface {
	Tick(tick int32) bool
	AnQ() []An
}

// BaseAn animates a dob
type BaseAn struct {
	Dob       *Dob
	anQ       []An
	duration  int64
	easer     Ease
	startTick int32
}

func (a *BaseAn) add(b An) An {
	if len(a.anQ) == 0 {
		a.anQ = make([]An, 0)
	}
	a.anQ = append(a.anQ, b)
	return b
}

// Tick never ends for BaseAn
func (a *BaseAn) Tick(tick int32) bool {
	return false
}

// PC returns percent complete
func (a *BaseAn) PC(tick int32) (raw float32, eased float32) {
	var pct float32
	if a.duration == 0 {
		pct = 1.0
	} else {
		pct = float32(tick-a.startTick) * float32(a.Dob.stage.DurationPerTick) / float32(a.duration)
		if pct > .99 {
			pct = 1
		}
	}
	return pct, a.easer(pct)
}

func (a *BaseAn) AnQ() []An {
	return a.anQ
}

// Move Anim
func (a *BaseAn) Move(x float32, y float32, duration time.Duration, easer Ease) *MoveAn {
	if easer == nil {
		easer = EaseNone
	}
	b := &MoveAn{BaseAn: BaseAn{Dob: a.Dob, anQ: nil, duration: int64(duration), easer: easer}, endX: x, endY: y}
	return a.add(b).(*MoveAn)
}

type MoveAn struct {
	BaseAn
	deltaX float32
	deltaY float32
	endX   float32
	endY   float32
}

func (a *MoveAn) Tick(tick int32) bool {
	if a.startTick == 0 {
		a.startTick = tick
		a.deltaX = a.endX - a.Dob.x
		a.deltaY = a.endY - a.Dob.y
	}
	pct, eased := a.PC(tick)
	a.Dob.x = a.endX - a.deltaX + eased*a.deltaX
	a.Dob.y = a.endY - a.deltaY + eased*a.deltaY
	return pct == 1
}

// Zoom Anim
type ZoomAn struct {
	BaseAn
	deltaZoom float32
	endZoom   float32
}

func (a *BaseAn) Zoom(zoom float32, duration time.Duration, easer Ease) *ZoomAn {
	if easer == nil {
		easer = EaseNone
	}
	b := &ZoomAn{BaseAn: BaseAn{Dob: a.Dob, anQ: nil, duration: int64(duration), easer: easer}, endZoom: zoom}
	return a.add(b).(*ZoomAn)
}

func (a *ZoomAn) Tick(tick int32) bool {
	if a.startTick == 0 {
		a.startTick = tick
		a.deltaZoom = a.endZoom - a.Dob.zoom
	}

	pct, eased := a.PC(tick)
	a.Dob.zoom = a.endZoom - a.deltaZoom + eased*a.deltaZoom

	return pct == 1
}

// Zoom Anim
type ExitAn struct {
	BaseAn
}

func (a *BaseAn) Exit() *ExitAn {
	b := &ExitAn{BaseAn: BaseAn{Dob: a.Dob, anQ: nil, duration: 0, easer: nil}}
	return a.add(b).(*ExitAn)
}

func (a *ExitAn) Tick(tick int32) bool {
	delete(a.Dob.stage.spawn, a.Dob.id)
	a.Dob.stage = nil
	return true
}

// View provides context for all DOBs
type View struct {
	H        int32
	Renderer *sdl.Renderer
	Stage    *Stage
	Title    string
	View     *sdl.Window
	W        int32
	sounds   map[string]*Wav
	textures map[string]*Tex
}

func (v *View) Init() (err error) {
	v.View, v.Renderer, err = sdl.CreateWindowAndRenderer(v.W, v.H, sdl.WINDOW_SHOWN)
	v.View.SetTitle(v.Title)
	v.textures = make(map[string]*Tex)
	v.sounds = make(map[string]*Wav)
	return
}

func (v *View) Destroy() {
	v.View.Destroy()
}

func (v *View) MakeStage() (s *Stage, err error) {
	s = &Stage{view: v}
	s.Dob.stage = s
	v.Stage = s
	return
}

func (v *View) TextureLoad(path string) (texture *Tex, err error) {
	var ok bool
	texture, ok = v.textures[path]
	if !ok {
		texture = &Tex{}
		texture.SDLTexture, err = img.LoadTexture(v.Renderer, path)
		if err != nil {
			err = fmt.Errorf("could not load texture at %s: %v", path, err)
			fmt.Println(err.Error())
			return nil, err
		}

		_, _, texture.W, texture.H, err = texture.SDLTexture.Query()
		if err != nil {
			err = fmt.Errorf("could not query texture at %s: %v", path, err)
			return nil, err
		}
	}
	return
}

func (v *View) SoundLoad(path string) (*Wav, error) {
	snd, ok := v.sounds[path]
	if !ok {
		wav, err := mix.LoadWAV(path)
		if err != nil {
			return nil, err
		}
		snd = &Wav{View: v, Wav: wav}
	}
	return snd, nil
}

// Stage is the root of the display tree
type Stage struct {
	Dob
	DurationPerTick int64
	view            *View
	SpawnId         int32
}

func (s *Stage) Tick(tick int32) {
	for _, v := range s.spawn {
		v.Tick(tick)
	}
}

func (s *Stage) Paint() {
	// clear
	err := s.view.Renderer.Clear()
	if err != nil {
		panic(err)
	}

	// paint dobs
	for _, v := range s.spawn {
		v.Paint()
	}

	s.view.Renderer.Present()
}

func (s *Stage) Play(fps int) {
	msPerFrame := int64(1000.0 / fps)
	s.DurationPerTick = msPerFrame * int64(time.Millisecond)

	// loop until the user quits
	running := true
	var tick int32 = 1
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
		sdl.Delay(uint32(msPerFrame))
		tick++
	}
}
