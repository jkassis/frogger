package gas

import (
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

// Texture wrapper mainly to cache the H and W of the texture
type Texture struct {
	SDLTexture *sdl.Texture
	H          int32
	W          int32
}

// Wav wrapper for sounds that provides a simple interface
// TODO consider getting rid of this abstraction for the native sdl.Wav
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
	w.playing = true
	channel, err := w.Wav.Play(0, -1)
	w.channel = channel
	return
}

func (w *Wav) Stop() {
	if !w.playing {
		return
	}
	w.playing = false
	mix.HaltChannel(w.channel)
}

// SDLC converts a uint32 to an sdl.Color
func SDLC(c uint32) sdl.Color {
	return sdl.Color{
		R: uint8(c >> 24),
		G: uint8(c << 8 >> 24),
		B: uint8(c << 16 >> 24),
		A: uint8(c << 24 >> 24),
	}
}
