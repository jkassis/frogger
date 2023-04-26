// Author: Jeremy Kassis (jkassis@gmail.com).
// Public domain software.
//
// A frogger game.

package main

import (
	"fmt"
	"frogger/gas"
	"math/rand"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func CHECK(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// main let's go
func main() {
	rand.Seed(time.Now().UnixNano())

	CHECK(gas.Init())
	defer gas.Destroy()

	view, err := gas.MakeView(800, 600, "Frogger")
	CHECK(err)
	defer view.Destroy()

	s, err := view.MakeStage()
	s.BGColor = sdl.Color{R: 0x01, G: 0xb3, B: 0x35, A: 0xff}
	// s.BGColor = sdl.Color{R: 196, G: 120, B: 161, A: 0xff}

	CHECK(err)

	montserrat96, _ := view.FontLoad("fonts/Montserrat-Regular.ttf", 96)
	montserrat48, _ := view.FontLoad("fonts/Montserrat-Regular.ttf", 48)
	CHECK(err)

	go func() {
		bg, _ := s.Root.Spawn("img/bg.png")
		bg.Move(400, 300, 0, nil)

		frog, _ := s.Root.Spawn("img/frog.png")
		frog.Scale = .05
		logo, _ := s.Root.Spawn("")
		logo.Scale = .1
		title, _ := s.Root.Spawn("")

		// title
		title.FillC(gas.SDLC(0x00ff00ff))
		title.TxtOutlineC(gas.SDLC(0x333333ff))
		title.TxtOutlineW = 4
		title.Text(montserrat96, "Frogger")
		title.Scale = .5
		title.Pos(800, 300).
			Move(400, 300, 2*time.Second, gas.EaseInOutSin)

		// frog
		frog.
			Pos(0, 200).
			Move(120, 300, 2*time.Second, gas.EaseInOutSin).
			Then(func(d *gas.Dob) {
				// move and zoom
				d.Move(300, 120, 2*time.Second, nil)
				d.Zoom(4, 2*time.Second, nil).
					Then(func(d *gas.Dob) {
						// move an zoom again. note how these race to Exit
						d.Zoom(.25, 3*time.Second, gas.EaseInOutSin).Exit()
						d.Move(330, 280, 3*time.Second, nil)
					})
			})

		// logo
		logo.FillC(gas.SDLC(0xffff33dd))
		logo.TxtOutlineC(gas.SDLC(0x003300dd))
		logo.TxtOutlineW = 2
		logo.Text(montserrat48, "jkassis ©2023")
		logo.
			Pos(50, 50).
			Move(120, 300, 2*time.Second, gas.EaseInOutSin).
			Then(func(d *gas.Dob) {
				d.Move(533, 400, 3*time.Second, gas.EaseInOutSin).
					Zoom(10, 3*time.Second, gas.EaseInOutSin).
					Then(func(d *gas.Dob) {
						// note how we trigger this title anim when the logo anim complete
						title.Zoom(2, 200*time.Millisecond, nil).
							Zoom(1, 400*time.Millisecond, nil)
					})
			})
		logo.Emit(logo, 20, 500*time.Millisecond, 3*time.Second, s.Root, gas.EaseInOutSinInv,
			func(d *gas.Dob) {
				x := float32(view.W) * rand.Float32()
				y := float32(view.H) * rand.Float32()
				dur := time.Second + time.Duration(rand.Float32()*float32(3*time.Second))
				d.Move(x, y, dur, nil).Exit()
			})
	}()

	s.Play(30)
}
