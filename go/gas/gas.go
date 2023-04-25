// gas is the game animation system
package gas

import (
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var anID int64
var dobID int64

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
	err = ttf.Init()
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
	fillC    sdl.Color // color
	outlineC sdl.Color
	OutlineW int
	D        [2]int32 // dim
	Px       float32  // posX
	Py       float32  // posY
	spawn    map[int64]*Dob
	spawner  *Dob
	Stage    *Stage
	Texture  *Tex
	zoom     float32
}

func (d *Dob) Tick(tick int32) {
	for ID, an := range d.anQ {
		if an.Tick(tick) {
			delete(d.anQ, ID)
			for nID, nAn := range an.AnQ() {
				d.anQ[nID] = nAn
			}
		}
	}
	for _, spawn := range d.spawn {
		spawn.Tick(tick)
	}
}

func (d *Dob) Paint() {
	dst := sdl.Rect{X: int32(d.Px - d.zoom*float32(d.D[0])/2), Y: int32(d.Py - d.zoom*float32(d.D[1])/2), W: int32(d.zoom * float32(d.D[0])), H: int32(d.zoom * float32(d.D[1]))}
	if d.Texture == nil {
		d.Stage.view.Renderer.SetDrawColor(d.fillC.R, d.fillC.G, d.fillC.B, d.fillC.A)
		d.Stage.view.Renderer.FillRect(&dst)
	} else {
		src := sdl.Rect{X: 0, Y: 0, W: d.D[0], H: d.D[1]}
		d.Stage.view.Renderer.Copy(d.Texture.SDLTexture, &src, &dst)
	}

	for _, spawn := range d.spawn {
		spawn.Paint()
	}
}

func (d *Dob) FillC(c uint32) {
	d.fillC = sdl.Color{
		R: uint8(c >> 24),
		G: uint8(c << 8 >> 24),
		B: uint8(c << 16 >> 24),
		A: uint8(c << 24 >> 24),
	}
}

func (d *Dob) OutlineC(c uint32) {
	d.outlineC = sdl.Color{
		R: uint8(c >> 24),
		G: uint8(c << 8 >> 24),
		B: uint8(c << 16 >> 24),
		A: uint8(c << 24 >> 24),
	}
}

func (d *Dob) Text(font *ttf.Font, t string) (err error) {
	if d.Texture != nil && d.Texture.SDLTexture != nil {
		d.Texture.SDLTexture.Destroy()
	}
	d.Texture = &Tex{}
	if d.OutlineW > 0 {
		font.SetOutline(d.OutlineW)
		outlineSurface, _ := font.RenderUTF8Blended(t, d.outlineC)
		font.SetOutline(0)
		fillSurface, _ := font.RenderUTF8Blended(t, d.fillC)
		src := &sdl.Rect{X: 0, Y: 0, W: fillSurface.W, H: fillSurface.H}
		dst := &sdl.Rect{X: int32(d.OutlineW), Y: int32(d.OutlineW), W: fillSurface.W, H: fillSurface.H}
		// fillSurface.SetBlendMode(sdl.BLENDMODE_BLEND)
		fillSurface.Blit(src, outlineSurface, dst)
		d.Texture.SDLTexture, _ = d.Stage.view.Renderer.CreateTextureFromSurface(outlineSurface)
		d.D[0] = outlineSurface.W
		d.D[1] = outlineSurface.H
		fillSurface.Free()
		outlineSurface.Free()
	} else {
		fillSurface, _ := font.RenderUTF8Solid(t, d.fillC)
		d.Texture.SDLTexture, _ = d.Stage.view.Renderer.CreateTextureFromSurface(fillSurface)
		d.D[0] = fillSurface.W
		d.D[1] = fillSurface.H
		fillSurface.Free()
	}
	return
}

func (d *Dob) Spawn(path string) (dob *Dob, err error) {
	dob = &Dob{
		BaseAn: BaseAn{id: dobID},
		Stage:  d.Stage,
		zoom:   1,
	}
	dobID++
	if path == "" {
		dob.D[0] = 2
		dob.D[1] = 2
		dob.fillC = sdl.Color{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	} else {
		dob.Texture, _ = d.Stage.view.TextureLoad(path)
		dob.D[0] = dob.Texture.W
		dob.D[1] = dob.Texture.H
	}
	dob.BaseAn.Dob = dob

	dob.spawner = d
	if d.spawn == nil {
		d.spawn = make(map[int64]*Dob, 0)
	}
	d.spawn[dob.id] = dob
	return
}

func (a *Dob) SpawnRm(b *Dob) {
	delete(a.spawn, b.id)
}

func (a *Dob) SpawnAdd(b *Dob) {
	a.spawn[b.id] = b
}

// An animates
type An interface {
	ID() int64
	Tick(tick int32) bool
	AnQ() map[int64]An
}

// BaseAn animates a dob
type BaseAn struct {
	id        int64
	Dob       *Dob
	anQ       map[int64]An
	Duration  int64
	Easer     Ease
	StartTick int32
}

func (a *BaseAn) ID() int64 {
	return a.id
}

func (a *BaseAn) add(b An) An {
	if len(a.anQ) == 0 {
		a.anQ = make(map[int64]An, 0)
	}
	a.anQ[b.ID()] = b
	return b
}

// Tick never ends for BaseAn
func (a *BaseAn) Tick(tick int32) bool {
	return false
}

// PC returns percent complete
func (a *BaseAn) PC(tick int32) (raw float32, eased float32) {
	var pct float32
	if a.Duration == 0 {
		pct = 1.0
	} else {
		pct = float32(tick-a.StartTick) * float32(a.Dob.Stage.DurationPerTick) / float32(a.Duration)
		if pct > .99 {
			pct = 1
		}
	}
	return pct, a.Easer(pct)
}

func (a *BaseAn) AnQ() map[int64]An {
	return a.anQ
}

// Move Anim
func (a *BaseAn) Move(x float32, y float32, duration time.Duration, easer Ease) *MoveAn {
	if easer == nil {
		easer = EaseNone
	}
	b := &MoveAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: int64(duration), Easer: easer}, endX: x, endY: y}
	anID++
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
	if a.StartTick == 0 {
		a.StartTick = tick
		a.deltaX = a.endX - a.Dob.Px
		a.deltaY = a.endY - a.Dob.Py
	}
	pct, eased := a.PC(tick)
	a.Dob.Px = a.endX - a.deltaX + eased*a.deltaX
	a.Dob.Py = a.endY - a.deltaY + eased*a.deltaY
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
	b := &ZoomAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: int64(duration), Easer: easer}, endZoom: zoom}
	anID++
	return a.add(b).(*ZoomAn)
}

func (a *ZoomAn) Tick(tick int32) bool {
	if a.StartTick == 0 {
		a.StartTick = tick
		a.deltaZoom = a.endZoom - a.Dob.zoom
	}

	pct, eased := a.PC(tick)
	a.Dob.zoom = a.endZoom - a.deltaZoom + eased*a.deltaZoom

	return pct == 1
}

// Spawn Anim
type EmitAn struct {
	BaseAn
	handler      func(*Dob)
	interval     time.Duration
	lastEmitTick int32
	qty          int
	target       *Dob
	template     *Dob
}

func (a *BaseAn) Emit(template *Dob, qty int, delayEach time.Duration, duration time.Duration, target *Dob, easer Ease, handler func(*Dob)) *EmitAn {
	if easer == nil {
		easer = EaseNone
	}
	b := &EmitAn{
		qty:      qty,
		BaseAn:   BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: int64(duration), Easer: easer},
		template: template,
		interval: delayEach,
		target:   target,
		handler:  handler,
	}
	anID++
	return a.add(b).(*EmitAn)
}

func (a *EmitAn) Tick(tick int32) bool {
	pct, _ := a.PC(tick)
	if a.StartTick == 0 || ((tick-a.lastEmitTick)*int32(a.Dob.Stage.DurationPerTick) > int32(a.interval)) {
		a.lastEmitTick = tick

		for i := 0; i < a.qty; i++ {
			c := a.template
			b, _ := c.Spawn("")
			b.fillC = c.fillC
			b.D = c.D
			b.Duration = c.Duration
			b.Easer = c.Easer
			b.Px = c.Px
			b.Py = c.Py
			b.Stage = c.Stage
			b.Texture = c.Texture
			b.zoom = c.zoom
			b.StartTick = 0

			if a.target != nil {
				delete(c.spawn, b.id)
				a.target.spawn[b.id] = b
				b.spawner = a.target
			}
			a.handler(b)
		}
	}
	if a.StartTick == 0 {
		a.StartTick = tick
	}
	return pct == 1
}

// Zoom Anim
type EndAn struct {
	BaseAn
}

func (a *BaseAn) End() *EndAn {
	b := &EndAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: 0, Easer: nil}}
	anID++
	return a.add(b).(*EndAn)
}

func (a *EndAn) Tick(tick int32) bool {
	delete(a.Dob.spawner.spawn, a.Dob.id)
	a.Dob.spawner = nil
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
	fonts    map[string]*ttf.Font
	textures map[string]*Tex
}

func (v *View) Init() (err error) {
	v.View, v.Renderer, err = sdl.CreateWindowAndRenderer(v.W, v.H, sdl.WINDOW_SHOWN)
	v.View.SetTitle(v.Title)
	v.textures = make(map[string]*Tex)
	v.sounds = make(map[string]*Wav)
	v.fonts = make(map[string]*ttf.Font)
	return
}

func (v *View) Destroy() {
	v.View.Destroy()
}

func (v *View) MakeStage() (s *Stage, err error) {
	s = &Stage{view: v}
	s.Root = &Dob{Stage: s, zoom: 1}
	s.Root.D[0] = v.W
	s.Root.D[1] = v.H
	s.Root.Px = float32(v.W / 2)
	s.Root.Py = float32(v.H / 2)
	s.Root.FillC(0xffffffff)
	dobID++
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

func (v *View) FontLoad(path string, size int) (font *ttf.Font, err error) {
	var ok bool
	font, ok = v.fonts[path+string(size)]
	if !ok {
		font, err = ttf.OpenFont(path, 48)
		if err != nil {
			return
		}
		v.fonts[path+string(size)] = font
	}
	return font, nil
}

// Stage is the root of the display tree
type Stage struct {
	DurationPerTick int64
	view            *View
	Root            *Dob
}

func (s *Stage) Play(fps int) {
	msPerFrame := int64(1000.0 / fps)
	s.DurationPerTick = msPerFrame * int64(time.Millisecond)

	// loop until the user quits
	running := true
	var tick int32 = 1
	for running {
		// wonder if performance of clear is better
		// s.view.Renderer.SetDrawColor(0xff, 0xff, 0xff, 0xff)
		// s.view.Renderer.Clear()
		s.Root.Tick(tick)
		s.Root.Paint()
		s.view.Renderer.Present()
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
