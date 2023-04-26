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

// Init initializes sdl dependencies for gas. Should be the first call when using the framework.
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

// Destroy quits sdl dependencies when clients no longer need gas. Call it with a defer after .Init
func Destroy() {
	sdl.Quit()
}

// View provides context for all DOBs (most notably the renderer)
type View struct {
	H        int32
	Renderer *sdl.Renderer
	Title    string
	W        int32
	fonts    map[string]*ttf.Font
	sounds   map[string]*Wav
	textures map[string]*Texture
	window   *sdl.Window
}

// MakeView  returns a gas.View which maps to and sdl window. Multiples ok.
func MakeView(w, h int32, title string) (view *View, err error) {
	view = &View{W: w, H: h, Title: "Frogger"}
	return view, view.Init()
}

// Init sets up a new view (HostOS window)
// TODO share textures, sounds, and fonts between views.
func (v *View) Init() (err error) {
	v.window, v.Renderer, err = sdl.CreateWindowAndRenderer(v.W, v.H, sdl.WINDOW_SHOWN)
	v.window.SetTitle(v.Title)
	v.textures = make(map[string]*Texture)
	v.sounds = make(map[string]*Wav)
	v.fonts = make(map[string]*ttf.Font)
	return
}

// Destroy releases the view window
func (v *View) Destroy() {
	v.window.Destroy()
}

func (v *View) TextureLoad(path string) (texture *Texture, err error) {
	var ok bool
	texture, ok = v.textures[path]
	if !ok {
		texture = &Texture{}
		texture.SDLTexture, err = img.LoadTexture(v.Renderer, path)
		if err != nil {
			err = fmt.Errorf("could not load texture at %s: %v", path, err)
			fmt.Println(err.Error())
			return nil, err
		}

		_, _, texture.W, texture.H, err = texture.SDLTexture.Query()
		if err != nil {
			err = fmt.Errorf("could not query texture at %s: %v", path, err)
			fmt.Println(err.Error())
			return nil, err
		}
		v.textures[path] = texture
	}
	return
}

func (v *View) SoundLoad(path string) (*Wav, error) {
	snd, ok := v.sounds[path]
	if !ok {
		wav, err := mix.LoadWAV(path)
		if err != nil {
			err = fmt.Errorf("could not load sound at %s: %v", path, err)
			return nil, err
		}
		snd = &Wav{View: v, Wav: wav}
		v.sounds[path] = snd
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

// MakeStage returns a new rendering context.
// TODO test that multiple stages work with one view.
func MakeStage(v *View) (s *Stage, err error) {
	s = &Stage{view: v}
	s.Root = &Dob{Stage: s, zoom: 1}
	s.Root.D[0] = v.W
	s.Root.D[1] = v.H
	s.Root.Px = float32(v.W / 2)
	s.Root.Py = float32(v.H / 2)
	s.Root.FillC = SDLC(0x00000000)
	s.Root.Scale = 1
	dobID++
	return
}

// Play starts the animation / rendering loop
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

// Dob (aka Display Object)
// Renders images, text, or colored boxes to the screen.
// Supports animation, z-layer nesting, etc.
type Dob struct {
	BaseAn
	D       [2]int32                    // dim
	FillC   sdl.Color                   // color to render if texture is nil
	Px      float32                     // posX
	Py      float32                     // posY
	Scale   float32                     // default scale of hi-rez text and graphics
	Stage   *Stage                      // provides access to context and renderer
	Texture *Texture                    // texture to render
	TxtOutC sdl.Color                   // color of the text outline
	TxtOutW int                         // outline width
	ctx     *Dob                        // the dob to which this dob is a child
	dobs    *maps.SliceMap[int64, *Dob] // children of this dob in the render order
	txt     string                      // actual text rendered in this dob
	txtFont *ttf.Font                   // text font
	zoom    float32                     // current zoom/scaling factor
}

// Tick
// Runs all Ans in the anSet and passes Tick down to embedded dobs
// Note: the tick is a number, not a time, making this deterministic
// were it not for the anSet iteration.
//
// TODO convert anSet to a SliceMap for determinism.
func (d *Dob) Tick(tick int32) {
	for ID, an := range d.anSet {
		if an.Tick(tick) {
			delete(d.anSet, ID)
			for nID, nAn := range an.AnSet() {
				d.anSet[nID] = nAn
			}
		}
	}

	d.dobs.Range(func(id int64, d *Dob) bool {
		// we need this check to support racing Ans
		if d != nil {
			d.Tick(tick)
		}
		return true
	})
}

// Paint
// Puts textures and rectangles on the view. Runs all all embedded dobs.
func (d *Dob) Paint() {
	dobsPainted++
	dst := sdl.Rect{X: int32(d.Px - d.Scale*d.zoom*float32(d.D[0])/2), Y: int32(d.Py - d.Scale*d.zoom*float32(d.D[1])/2), W: int32(d.Scale * d.zoom * float32(d.D[0])), H: int32(d.Scale * d.zoom * float32(d.D[1]))}
	if d.Texture != nil {
		src := sdl.Rect{X: 0, Y: 0, W: d.D[0], H: d.D[1]}
		d.Stage.view.Renderer.Copy(d.Texture.SDLTexture, &src, &dst)
	} else if d.FillC.A > 0 {
		d.Stage.view.Renderer.SetDrawColor(d.FillC.R, d.FillC.G, d.FillC.B, d.FillC.A)
		d.Stage.view.Renderer.FillRect(&dst)
	}

	d.dobs.Range(func(id int64, d *Dob) bool {
		// we have to do this check to support racing Ans
		if d != nil {
			d.Paint()
		}
		return true
	})
}

// TxtOut sugar to set text outline properties all at once
func (d *Dob) TxtOut(outW int, outC sdl.Color) {
	d.TxtOutC = outC
	d.TxtOutW = outW
}

// TxtFill sugar to set TxtFill proerties all at once
func (d *Dob) TxtFill(txt string, fillC sdl.Color, font *ttf.Font) {
	d.FillC = fillC
	d.txt = txt
	d.txtFont = font
}

// TxtFillOut sugar to set all text properties all at once and render
func (d *Dob) TxtFillOut(txt string, fillC sdl.Color, font *ttf.Font, outW int, outC sdl.Color) {
	d.FillC = fillC
	d.TxtOutC = outC
	d.TxtOutW = outW
	d.txt = txt
	d.txtFont = font
	d.TxtRender()
}

// TxtRender renders text. Call after changes to text properties.
func (d *Dob) TxtRender() (err error) {
	if d.Texture != nil && d.Texture.SDLTexture != nil {
		d.Texture.SDLTexture.Destroy()
	}
	d.Texture = &Texture{}
	if d.TxtOutW > 0 {
		// render text with outline
		d.txtFont.SetOutline(d.TxtOutW)
		outlineSurface, _ := d.txtFont.RenderUTF8Blended(d.txt, d.TxtOutC)
		d.txtFont.SetOutline(0)
		fillSurface, _ := d.txtFont.RenderUTF8Blended(d.txt, d.FillC)
		src := &sdl.Rect{X: 0, Y: 0, W: fillSurface.W, H: fillSurface.H}
		dst := &sdl.Rect{X: int32(d.TxtOutW), Y: int32(d.TxtOutW), W: fillSurface.W, H: fillSurface.H}
		// fillSurface.SetBlendMode(sdl.BLENDMODE_BLEND)
		fillSurface.Blit(src, outlineSurface, dst)
		d.Texture.SDLTexture, _ = d.Stage.view.Renderer.CreateTextureFromSurface(outlineSurface)
		d.D[0] = outlineSurface.W
		d.D[1] = outlineSurface.H
		fillSurface.Free()
		outlineSurface.Free()
	} else {
		// render text without outline
		fillSurface, _ := d.txtFont.RenderUTF8Solid(d.txt, d.FillC)
		d.Texture.SDLTexture, _ = d.Stage.view.Renderer.CreateTextureFromSurface(fillSurface)
		d.D[0] = fillSurface.W
		d.D[1] = fillSurface.H
		fillSurface.Free()
	}
	return
}

// Spawn yields a new dob with d as its ctx.
// This implies a parent-child relationship, which only affects render order now.
// In future versions, dobs might use their ctx dobs as a reference frame for
// relative positioning, zooming, etc.
func (d *Dob) Spawn(path string) (dob *Dob, err error) {
	dobID++
	dob = &Dob{
		BaseAn: BaseAn{id: dobID},
		Stage:  d.Stage,
		zoom:   1,
		Scale:  d.Scale,
	}
	if path == "" {
		// spawn a color rectangle
		dob.D[0] = 2
		dob.D[1] = 2
		dob.FillC = sdl.Color{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	} else {
		// spawn a texture
		dob.Texture, _ = d.Stage.view.TextureLoad(path)
		dob.D[0] = dob.Texture.W
		dob.D[1] = dob.Texture.H
	}
	dob.BaseAn.dob = dob

	// add to the spawner
	d.DobAdd(dob)
	return
}

// DobAdd adds b to d
func (d *Dob) DobAdd(b *Dob) {
	if b.ctx != nil { // dobs can only link into the tree once
		b.ctx.DobRm(b)
	}
	if d.dobs == nil {
		d.dobs = &maps.SliceMap[int64, *Dob]{}
	}
	d.dobs.Set(b.id, b)
	b.ctx = d
}

// DobRm this orphans b unless client code holds a reference
func (d *Dob) DobRm(b *Dob) {
	d.dobs.Delete(b.id)
	b.ctx = nil
}

// DobsClear removes all dobs. You probably want to call AnSetClear too.
func (d *Dob) DobsClear() {
	d.dobs.Clear()
}

// AnSetClear empties the AnSet. You probably want to call DobsClear too.
func (d *Dob) AnSetClear() {
	d.anSet = nil
}

// An is the dob animation interface
type An interface {
	AnSet() map[int64]An
	Dob() *Dob
	ID() int64
	Tick(tick int32) bool
}

// BaseAn animates a dob
type BaseAn struct {
	id        int64        // unique id for the animation (and the Dob since all dobs embed BaseAn)
	dob       *Dob         // the target of animation
	anSet     map[int64]An // Set of simultaneously running animations mutating the dob state
	Duration  int64        // duration of the animation, after which it is over and removed from the anSet
	Easer     Ease         // applies easing the the rate of the animation
	StartTick int32        // first value of Tick passed to An.Tick // TODO set this before entry.
}

// Dob returns the dob target
func (a *BaseAn) Dob() *Dob {
	return a.dob
}

// ID returns the id for the An.
// As discussed... when called on the BaseAn of a dob, this represents Dob.id
func (a *BaseAn) ID() int64 {
	return a.id
}

// AnSetAdd activates the animation.
// TODO validate that the added an runs on the *current* tick. Otherwise a 1sec an followed by
// another 1sec an will take 2sec + frameTimeMs to complete.
// check how anSet and SliceMap.Range iterate
func (a *BaseAn) AnSetAdd(b An) An {
	if a.anSet == nil {
		a.anSet = make(map[int64]An, 0)
	}
	a.anSet[b.ID()] = b
	return b
}

// PC calculates raw percent complete and the and eased percent complete
func (a *BaseAn) PC(tick int32) (raw float32, eased float32) {
	var pct float32
	if a.Duration == 0 {
		pct = 1.0
	} else {
		pct = float32(tick-a.StartTick) * float32(a.dob.Stage.DurationPerTick) / float32(a.Duration)
		if pct > .99 {
			pct = 1
		}
	}
	return pct, a.Easer(pct)
}

// AnSet gets the anSet
func (a *BaseAn) AnSet() map[int64]An {
	return a.anSet
}

// Move sets the position, but does not create a new An. returns the last an added
func (a *BaseAn) Move(x float32, y float32) *BaseAn {
	a.dob.Px = x
	a.dob.Py = y
	return a
}

// MoveToAn moves a dob over time with easing
type MoveToAn struct {
	BaseAn
	deltaX float32
	deltaY float32
	endX   float32
	endY   float32
}

// MoveTo yields a MoveAn for BaseAn.Dob
func (a *BaseAn) MoveTo(x float32, y float32, duration time.Duration, easer Ease) *MoveToAn {
	anID++
	if easer == nil {
		easer = EaseNone
	}
	b := &MoveToAn{BaseAn: BaseAn{id: anID, dob: a.dob, anSet: nil, Duration: int64(duration), Easer: easer}, endX: x, endY: y}
	return a.AnSetAdd(b).(*MoveToAn)
}

func (a *MoveToAn) Tick(tick int32) bool {
	if a.StartTick == 0 {
		a.StartTick = tick
		a.deltaX = a.endX - a.dob.Px
		a.deltaY = a.endY - a.dob.Py
	}
	pct, eased := a.PC(tick)
	a.dob.Px = a.endX - a.deltaX + eased*a.deltaX
	a.dob.Py = a.endY - a.deltaY + eased*a.deltaY
	return pct == 1
}

// Zoom sets the zoom, but does not create a new An. returns the last an added
func (a *BaseAn) Zoom(z float32) *BaseAn {
	a.dob.zoom = z
	return a
}

// ZoomAn animates the zoom factor for a dob with easing
type ZoomAn struct {
	BaseAn
	deltaZoom float32
	endZoom   float32
}

// ZoomTo yields a ZoomAn for BaseAn.Dob
func (a *BaseAn) ZoomTo(zoom float32, duration time.Duration, easer Ease) *ZoomAn {
	anID++
	if easer == nil {
		easer = EaseNone
	}
	b := &ZoomAn{BaseAn: BaseAn{id: anID, dob: a.dob, anSet: nil, Duration: int64(duration), Easer: easer}, endZoom: zoom}
	return a.AnSetAdd(b).(*ZoomAn)
}

func (a *ZoomAn) Tick(tick int32) bool {
	if a.StartTick == 0 {
		a.StartTick = tick
		a.deltaZoom = a.endZoom - a.dob.zoom
	}

	pct, eased := a.PC(tick)
	a.dob.zoom = a.endZoom - a.deltaZoom + eased*a.deltaZoom

	return pct == 1
}

// EmitAn spawns qty dobs emitter's position every interval for a duration
// The lower bound of interval is the frame rate of the stage
// EmitAn calls "then" for each emitted Dob. Use "then" to start Ans on the emitted dob
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
		BaseAn:   BaseAn{id: anID, dob: a.dob, anSet: nil, Duration: int64(duration), Easer: easer},
		template: template,
		interval: delayEach,
		target:   target,
		Then:     handler,
	}
	return a.AnSetAdd(b).(*EmitAn)
}

func (a *EmitAn) Tick(tick int32) bool {
	if a.StartTick == 0 || ((tick-a.lastEmitTick)*int32(a.dob.Stage.DurationPerTick) > int32(a.interval)) {
		a.lastEmitTick = tick

		for i := 0; i < a.qty; i++ {
			c := a.template
			b, _ := c.Spawn("")
			b.FillC = c.FillC
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
	pct, _ := a.PC(tick)
	return pct == 1
}

// ThenAn calls the "then" function when it Ticks.
// It completes immediately so that chained Ans get added to the active set.
type ThenAn struct {
	BaseAn
	then func(*Dob)
}

// Then yields a ThenAn for BaseAn.Dob
func (a *BaseAn) Then(fn func(*Dob)) *ThenAn {
	anID++
	b := &ThenAn{BaseAn: BaseAn{id: anID, dob: a.dob, anSet: nil, Duration: 0, Easer: nil}, then: fn}
	return a.AnSetAdd(b).(*ThenAn)
}

func (a *ThenAn) Tick(tick int32) bool {
	a.then(a.dob)
	return true
}

// ThenWaitAn calls the "then" function when it Ticks.
// The "then" function must return an unbuffered channel.
// A ThenWaitAn starts a go routine to wait for the channel to close.
// When the channel closes, ThenWaitAn completes.
// This allows the animator to pause a chain to run others animations in parallel
// or do general computation.
//
// Note that this requires a goroutine, so watch resource utilization if you emit
// lots of particles that use this.
type ThenWaitAn struct {
	BaseAn
	then     func(*Dob) chan struct{}
	complete bool
}

// ThenWait yields a ThenWaitAn for BaseAn.Dob
func (a *BaseAn) ThenWait(fn func(*Dob) chan struct{}) *ThenAn {
	anID++
	b := &ThenWaitAn{BaseAn: BaseAn{id: anID, dob: a.dob, anSet: nil, Duration: 0, Easer: nil}, then: fn}
	return a.AnSetAdd(b).(*ThenAn)
}

func (a *ThenWaitAn) Tick(tick int32) bool {
	// we wait with a goroutine so that the main animation loop can continue
	go func() {
		<-a.then(a.dob)
		a.complete = true
	}()
	return a.complete
}

// ExitAn removes a dob from the screen / stage ("exit stage left!")
// If the spawning code has no reference, the garbage collector will get the dob.
type ExitAn struct {
	BaseAn
}

// Exit yields an ExitAn for BaseAn.Dob
func (a *BaseAn) Exit() *ExitAn {
	anID++
	b := &ExitAn{BaseAn: BaseAn{id: anID, dob: a.dob, anSet: nil, Duration: 0, Easer: nil}}
	return a.AnSetAdd(b).(*ExitAn)
}

func (a *ExitAn) Tick(tick int32) bool {
	a.dob.ctx.DobRm(a.dob)
	return true
}
