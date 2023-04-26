// gas is the game animation system
package gas

import (
	"fmt"
	"time"

	"github.com/goradd/maps"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var anID int64
var dobID int64
var dobsPainted int64

func Init() (err error) {
	// init sdl
	err = sdl.Init(sdl.INIT_AUDIO | sdl.INIT_EVENTS | sdl.INIT_TIMER | sdl.INIT_VIDEO)
	if err != nil {
		return err
	}
	err = img.Init(img.INIT_PNG | img.INIT_JPG | img.INIT_WEBP)
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
	D           [2]int32                    // dim
	Px          float32                     // posX
	Py          float32                     // posY
	Scale       float32                     // default scale of hi-rez text and graphics
	Stage       *Stage                      // provides access to context and renderer
	Texture     *Tex                        // texture to render
	TxtOutlineW int                         // outline width
	ctx         *Dob                        // the dob to which this dob is a child
	dobs        *maps.SliceMap[int64, *Dob] // children of this dob in the render order
	fillC       sdl.Color                   // color to render if texture is nil
	txtOutlineC sdl.Color                   // color of the text outline
	zoom        float32                     // current zoom/scaling factor
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

	d.dobs.Range(func(id int64, d *Dob) bool {
		// we have to do this check to support racing anims
		if d != nil {
			d.Tick(tick)
		}
		return true
	})
}

func (d *Dob) Paint() {
	dobsPainted++
	dst := sdl.Rect{X: int32(d.Px - d.Scale*d.zoom*float32(d.D[0])/2), Y: int32(d.Py - d.Scale*d.zoom*float32(d.D[1])/2), W: int32(d.Scale * d.zoom * float32(d.D[0])), H: int32(d.Scale * d.zoom * float32(d.D[1]))}
	if d.Texture != nil {
		src := sdl.Rect{X: 0, Y: 0, W: d.D[0], H: d.D[1]}
		d.Stage.view.Renderer.Copy(d.Texture.SDLTexture, &src, &dst)
	} else if d.fillC.A > 0 {
		d.Stage.view.Renderer.SetDrawColor(d.fillC.R, d.fillC.G, d.fillC.B, d.fillC.A)
		d.Stage.view.Renderer.FillRect(&dst)
	}
	d.dobs.Range(func(id int64, d *Dob) bool {
		// we have to do this check to support racing anims
		if d != nil {
			d.Paint()
		}
		return true
	})
}

func (d *Dob) FillC(c sdl.Color) {
	// TODO re-render text
	d.fillC = c
}

func (d *Dob) TxtOutlineC(c sdl.Color) {
	// TODO re-render text
	d.txtOutlineC = c
}

func (d *Dob) Text(font *ttf.Font, t string) (err error) {
	if d.Texture != nil && d.Texture.SDLTexture != nil {
		d.Texture.SDLTexture.Destroy()
	}
	d.Texture = &Tex{}
	if d.TxtOutlineW > 0 {
		font.SetOutline(d.TxtOutlineW)
		outlineSurface, _ := font.RenderUTF8Blended(t, d.txtOutlineC)
		font.SetOutline(0)
		fillSurface, _ := font.RenderUTF8Blended(t, d.fillC)
		src := &sdl.Rect{X: 0, Y: 0, W: fillSurface.W, H: fillSurface.H}
		dst := &sdl.Rect{X: int32(d.TxtOutlineW), Y: int32(d.TxtOutlineW), W: fillSurface.W, H: fillSurface.H}
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

// Spawn yields a new dob with d as its ctx. This implies a parentt-child
// relationship, which currently only affects render order. In future versions,
// dobs might use their ctx dobs as a reference frame for relative positioning,
// zooming, etc.
func (d *Dob) Spawn(path string) (dob *Dob, err error) {
	dobID++
	dob = &Dob{
		BaseAn: BaseAn{id: dobID},
		Stage:  d.Stage,
		zoom:   1,
		Scale:  d.Scale,
	}
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

	d.DobAdd(dob)
	return
}

func (d *Dob) DobRm(b *Dob) {
	d.dobs.Delete(b.id)
	b.ctx = nil
}

func (d *Dob) DobAdd(b *Dob) {
	if d.dobs == nil {
		d.dobs = &maps.SliceMap[int64, *Dob]{}
	}
	d.dobs.Set(b.id, b)
	b.ctx = d
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
	if a.anQ == nil {
		a.anQ = make(map[int64]An, 0)
	}
	a.anQ[b.ID()] = b
	return b
}

// Tick never ends for BaseAn
func (a *BaseAn) Tick(tick int32) bool {
	return false
}

// PC calculates raw and eased percent complete
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

// AnQ gets the anQ
func (a *BaseAn) AnQ() map[int64]An {
	return a.anQ
}

// Pos just sets position
func (a *BaseAn) Pos(x float32, y float32) *BaseAn {
	a.Dob.Px = x
	a.Dob.Py = y
	return a
}

// MoveAn moves a dob to a point over time with easing
type MoveAn struct {
	BaseAn
	deltaX float32
	deltaY float32
	endX   float32
	endY   float32
}

// Move yields a MoveAn for BaseAn.Dob
func (a *BaseAn) Move(x float32, y float32, duration time.Duration, easer Ease) *MoveAn {
	anID++
	if easer == nil {
		easer = EaseNone
	}
	b := &MoveAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: int64(duration), Easer: easer}, endX: x, endY: y}
	return a.add(b).(*MoveAn)
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

// ZoomAn animates the zoom factor for a dob with easing
type ZoomAn struct {
	BaseAn
	deltaZoom float32
	endZoom   float32
}

// Zoom yields a ZoomAn for BaseAn.Dob
func (a *BaseAn) Zoom(zoom float32, duration time.Duration, easer Ease) *ZoomAn {
	anID++
	if easer == nil {
		easer = EaseNone
	}
	b := &ZoomAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: int64(duration), Easer: easer}, endZoom: zoom}
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

// EmitAn spawns qty dobs at the current dobs location every interval for a duration
// the lower limit of interval is the frame rate of the stage
// EmitAn calls "then" for each emitted Dob. Use "then" to start Ans on emitted dobs.
type EmitAn struct {
	BaseAn
	Then         func(*Dob)
	interval     time.Duration
	lastEmitTick int32
	qty          int
	target       *Dob
	template     *Dob
}

// Emit yields an EmitAn for BaseAn.Dob
func (a *BaseAn) Emit(template *Dob, qty int, delayEach time.Duration, duration time.Duration, target *Dob, easer Ease, handler func(*Dob)) *EmitAn {
	anID++
	if easer == nil {
		easer = EaseNone
	}
	b := &EmitAn{
		qty:      qty,
		BaseAn:   BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: int64(duration), Easer: easer},
		template: template,
		interval: delayEach,
		target:   target,
		Then:     handler,
	}
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
				c.dobs.Delete(b.id)
				a.target.dobs.Set(b.id, b)
				b.ctx = a.target
			}
			a.Then(b)
		}
	}
	if a.StartTick == 0 {
		a.StartTick = tick
	}
	return pct == 1
}

// ThenAn calls the "then" function when all Ans of the preceding chain complete
// It completes immediately and does not wait.
type ThenAn struct {
	BaseAn
	then func(*Dob)
}

// Then yields a ThenAn for BaseAn.Dob
func (a *BaseAn) Then(fn func(*Dob)) *ThenAn {
	anID++
	b := &ThenAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: 0, Easer: nil}, then: fn}
	return a.add(b).(*ThenAn)
}

func (a *ThenAn) Tick(tick int32) bool {
	a.then(a.Dob)
	return true
}

// ThenWaitAn calls the "then" function when all Ans of the preceding chain complete
// The "then" function should return an unbuffered channel and close it when then should complete.
// This allows the then function to pause the main chain while it completes.
// Note that this requires use of a goroutine, so watch your resource utilization if you
// emit lots of particles that rely on this.
type ThenWaitAn struct {
	BaseAn
	then     func(*Dob) chan struct{}
	complete bool
}

// ThenWait yields a ThenWaitAn for BaseAn.Dob
func (a *BaseAn) ThenWait(fn func(*Dob) chan struct{}) *ThenAn {
	anID++
	b := &ThenWaitAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: 0, Easer: nil}, then: fn}
	return a.add(b).(*ThenAn)
}

func (a *ThenWaitAn) Tick(tick int32) bool {
	go func() {
		<-a.then(a.Dob)
		a.complete = true
	}()
	return a.complete
}

// ExitAn removes a dob from the screen / stage
// if the spawning code has no reference, it will get garbage collected
type ExitAn struct {
	BaseAn
}

// Exit yields an ExitAn for BaseAn.Dob
func (a *BaseAn) Exit() *ExitAn {
	anID++
	b := &ExitAn{BaseAn: BaseAn{id: anID, Dob: a.Dob, anQ: nil, Duration: 0, Easer: nil}}
	return a.add(b).(*ExitAn)
}

func (a *ExitAn) Tick(tick int32) bool {
	a.Dob.ctx.DobRm(a.Dob)
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
	fonts    map[string]*ttf.Font
	sounds   map[string]*Wav
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
	s.Root.FillC(SDLC(0x00000000))
	s.Root.Scale = 1
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
	key := fmt.Sprintf("%s-%d", path, size)
	var ok bool
	font, ok = v.fonts[key]
	if !ok {
		font, err = ttf.OpenFont(path, size)
		if err != nil {
			return
		}
		v.fonts[key] = font
	}
	return font, nil
}

// Stage is the root of the display tree
type Stage struct {
	DurationPerTick int64
	view            *View
	BGColor         sdl.Color
	Root            *Dob
	logTickLast     int32
}

func (s *Stage) Play(fps int) {
	msPerFrame := int64(1000.0 / fps)
	s.DurationPerTick = msPerFrame * int64(time.Millisecond)

	// loop until the user quits
	running := true
	var tick int32 = 1
	for running {
		dobsPainted = 0
		s.view.Renderer.SetDrawColor(s.BGColor.R, s.BGColor.G, s.BGColor.B, s.BGColor.A)
		s.view.Renderer.Clear()
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
		if int(tick-s.logTickLast) >= fps {
			s.logTickLast = tick
			fmt.Printf("dobs painted: %d\n", dobsPainted)
		}
		tick++
	}
}

func SDLC(c uint32) sdl.Color {
	return sdl.Color{
		R: uint8(c >> 24),
		G: uint8(c << 8 >> 24),
		B: uint8(c << 16 >> 24),
		A: uint8(c << 24 >> 24),
	}
}
