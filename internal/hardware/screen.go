package hardware

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/trondhumbor/chip8/internal/config"
	"golang.org/x/image/colornames"
)

const scaleFactor = 20

func NewScreen(cfg config.Config, screenChan <-chan Screenbuffer, keyboardChan chan<- byte) {
	// Temporary empty buffer until we receive one from the CPU
	screenBuffer := *NewScreenbuffer(cfg.ScreenSizeX, cfg.ScreenSizeY)

	pixelCfg := pixelgl.WindowConfig{
		Title:  "CHIP8",
		Bounds: pixel.R(0, 0, float64(cfg.ScreenSizeX*scaleFactor), float64(cfg.ScreenSizeY*scaleFactor)),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(pixelCfg)
	if err != nil {
		panic(err)
	}

	keyMapping := map[pixelgl.Button]byte{
		pixelgl.Key0:   0x0,
		pixelgl.Key1:   0x1,
		pixelgl.Key2:   0x2,
		pixelgl.Key3:   0x3,
		pixelgl.Key4:   0x4,
		pixelgl.Key5:   0x5,
		pixelgl.Key6:   0x6,
		pixelgl.Key7:   0x7,
		pixelgl.Key8:   0x8,
		pixelgl.Key9:   0x9,
		pixelgl.KeyKP0: 0x0,
		pixelgl.KeyKP1: 0x1,
		pixelgl.KeyKP2: 0x2,
		pixelgl.KeyKP3: 0x3,
		pixelgl.KeyKP4: 0x4,
		pixelgl.KeyKP5: 0x5,
		pixelgl.KeyKP6: 0x6,
		pixelgl.KeyKP7: 0x7,
		pixelgl.KeyKP8: 0x8,
		pixelgl.KeyKP9: 0x9,
		pixelgl.KeyA:   0xA,
		pixelgl.KeyB:   0xB,
		pixelgl.KeyC:   0xC,
		pixelgl.KeyD:   0xD,
		pixelgl.KeyE:   0xE,
		pixelgl.KeyF:   0xF,
	}

	waitForKeypress := func() {
		for {
			for k, v := range keyMapping {
				if win.Pressed(k) {
					// write this to the keylogger channel
					keyboardChan <- v
				}
			}
		}
	}

	updateScreenbuffer := func() {
		for {
			select {
			case sc := <-screenChan:
				screenBuffer = sc
			}
		}
	}

	go waitForKeypress()
	go updateScreenbuffer()

	win.Clear(colornames.Black)
	imd := imdraw.New(nil)

	var (
		frames = 0
		second = time.Tick(time.Second)
	)

	for !win.Closed() {
		imd.Clear()

		for y := 0; y < cfg.ScreenSizeY; y++ {
			for x := 0; x < cfg.ScreenSizeX; x++ {
				if screenBuffer.Get(x, y) == 1 {
					imd.Color = colornames.White
					xVec := pixel.V(float64(x), float64(cfg.ScreenSizeY-y)).Scaled(scaleFactor)
					yVec := pixel.V(float64(x)+1, float64(cfg.ScreenSizeY-y)-1).Scaled(scaleFactor)
					imd.Push(xVec, yVec)
					imd.Rectangle(0)
				}
			}
		}

		win.Clear(colornames.Black)
		imd.Draw(win)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", pixelCfg.Title, frames))
			frames = 0
		default:
		}
	}
}
