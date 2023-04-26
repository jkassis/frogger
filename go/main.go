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
	CHECK(err)

	montserrat, err := view.FontLoad("fonts/Montserrat-Regular.ttf", 48)
	CHECK(err)

	go func() {
		frog, _ := s.Root.Spawn("img/frog.png")
		frog.ZoomBase = .05
		logo, _ := s.Root.Spawn("")
		logo.ZoomBase = .1
		title, _ := s.Root.Spawn("")

		// title
		title.FillC(0x00ff00ff)
		title.OutlineC(0x333333ff)
		title.OutlineW = 2
		title.Text(montserrat, "Frogger")
		title.Move(800, 300, 0, nil)
		title.Move(400, 300, 2*time.Second, gas.EaseInOutSin)

		// frogA
		frogA := frog.
			Move(120, 300, 2*time.Second, gas.EaseInOutSin)

		frogA.Move(300, 120, 2*time.Second, nil)
		frogA.Zoom(4, 3*time.Second, nil).
			Zoom(1, 3*time.Second, gas.EaseInOutSin).
			End()

		// logo
		logo.FillC(0xffff33dd)
		logo.OutlineC(0x003300dd)
		logo.OutlineW = 2
		logo.Text(montserrat, "jkassis")
		logo.Emit(logo, 20, 500*time.Millisecond, 3*time.Second, s.Root, gas.EaseInOutSinInv, func(d *gas.Dob) {
			x := float32(view.W) * rand.Float32()
			y := float32(view.H) * rand.Float32()
			dur := time.Second + time.Duration(rand.Float32()*float32(3*time.Second))
			d.Move(x, y, dur, nil).End()
		})
		b := logo.
			Move(50, 50, 0, nil).
			Move(120, 300, 2*time.Second, gas.EaseInOutSin)

		b.Move(600, 400, 3*time.Second, gas.EaseInOutSin).
			Zoom(10, 3*time.Second, gas.EaseInOutSin).
			Then(func() {
				title.Zoom(2, 200*time.Millisecond, nil).
					Zoom(1, 400*time.Millisecond, nil)
			})

	}()

	s.Play(30)
}
