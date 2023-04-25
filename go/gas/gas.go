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
	stage   *Stage
	h       int32
	spawn   []*Dob
	texture *Tex
	w       int32
	x       float64
	y       float64
}

func (d *Dob) Tick(tick int64) {
	for j, anim := range d.Anims {
		if anim.Tick(tick) {
			d.Anims = append(d.Anims[:j], d.Anims[j+1:]...)
		}
	}
}

func (d *Dob) Paint() {
	src := sdl.Rect{X: 0, Y: 0, W: d.w * 2, H: d.h * 2}
	dst := sdl.Rect{X: int32(d.x), Y: int32(d.y), W: d.w, H: d.h}
	d.stage.view.Renderer.Copy(d.texture.SDLTexture, &src, &dst)
}

func (d *Dob) Spawn(path string) (dob *Dob, err error) {
	dob = &Dob{stage: d.stage}
	dob.texture, _ = d.stage.view.TextureLoad(path)
	dob.h = dob.texture.H
	dob.w = dob.texture.W
	dob.BaseAn.Dob = dob
	if dob.spawn == nil {
		dob.spawn = make([]*Dob, 0)
	}
	d.spawn = append(d.spawn, dob)
	return
}

// An animates
type An interface {
	Tick(tick int64) bool
}

// BaseAn animates a dob
type BaseAn struct {
	Dob   *Dob
	Anims []An
}

func (b *BaseAn) anim(a An) An {
	if len(b.Anims) == 0 {
		b.Anims = make([]An, 0)
	}
	b.Anims = append(b.Anims, a)
	return a
}

func (b *BaseAn) Move(x float64, y float64, duration time.Duration, easer Ease) An {
	if easer == nil {
		easer = EaseNone
	}
	moveAnim := &MoveAn{dob: b.Dob, endX: x, endY: y, duration: int64(duration), easer: easer}
	return b.anim(moveAnim)
}

func (m *BaseAn) Tick(tick int64) {
}

// Move Anim
type MoveAn struct {
	BaseAn
	easer     Ease
	deltaX    float64
	deltaY    float64
	dob       *Dob
	duration  int64
	endX      float64
	endY      float64
	startTick int64
}

func (m *MoveAn) Tick(tick int64) bool {
	if m.startTick == 0 {
		m.startTick = tick
		m.deltaX = m.endX - m.dob.x
		m.deltaY = m.endY - m.dob.y
	}
	pct := float64(tick-m.startTick) * float64(m.dob.stage.DurationPerTick) / float64(m.duration)
	if pct > 1 {
		pct = 1
	}
	pctWEase := m.easer(pct)
	m.dob.x = m.endX - m.deltaX + pctWEase*m.deltaX
	m.dob.y = m.endY - m.deltaY + pctWEase*m.deltaY
	return pct == 1
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
}

func (s *Stage) Tick(tick int64) {
	for i := 0; i < len(s.spawn); i++ {
		s.spawn[i].Tick(tick)
	}
}

func (s *Stage) Paint() {
	// clear
	err := s.view.Renderer.Clear()
	if err != nil {
		panic(err)
	}

	// paint dobs
	for i := 0; i < len(s.spawn); i++ {
		s.spawn[i].Paint()
	}

	s.view.Renderer.Present()
}

func (s *Stage) Play(fps int) {
	msPerFrame := uint32(1000.0 / fps)
	s.DurationPerTick = int64(msPerFrame) * int64(time.Millisecond)

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
		sdl.Delay(msPerFrame)
		tick++
	}
}
