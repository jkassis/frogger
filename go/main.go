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

	font, err := view.FontLoad("fonts/Montserrat-Regular.ttf", 48)
	CHECK(err)

	go func() {
		greenBlock, _ := s.Root.Spawn("img/block_green.png")

		// move then zoom
		a := greenBlock.
			Move(120, 300, 3*time.Second, gas.EaseInOutSin).
			Zoom(2, 0, nil)

			// move and zoom
		a.Move(300, 120, 3*time.Second, nil)
		a.Zoom(4, 3*time.Second, nil).
			Zoom(1, 3*time.Second, gas.EaseInOutSin).
			End()

		// rainbow block
		rainbowBlock, _ := s.Root.Spawn("")
		rainbowBlock.Color(0xff3333dd)
		rainbowBlock.Emit(rainbowBlock, 20, 500*time.Millisecond, 3*time.Second, s.Root, gas.EaseInOutSinInv, func(d *gas.Dob) {
			x := float32(view.W) * rand.Float32()
			y := float32(view.H) * rand.Float32()
			dur := time.Second + time.Duration(rand.Float32()*float32(3*time.Second))
			d.Move(x, y, dur, nil).End()
		})
		b := rainbowBlock.
			Move(50, 50, 0, nil).
			Move(120, 300, 3*time.Second, gas.EaseInOutSin)

		b.Move(600, 400, 3*time.Second, gas.EaseInOutSin).
			Zoom(10, 3*time.Second, gas.EaseInOutSin)

		frogger, _ := s.Root.Spawn("")
		frogger.Color(0x00ff33)
		frogger.Text(font, "Frogger")
		frogger.Move(800, 300, 0, nil)
		frogger.Move(400, 300, 3*time.Second, gas.EaseInOutSin)
	}()

	s.Play(30)
}
