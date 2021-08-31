package emulator

import (
	"encoding/binary"
	"os"

	"github.com/faiface/pixel/pixelgl"
	"github.com/trondhumbor/chip8/internal/config"
	"github.com/trondhumbor/chip8/internal/hardware"
)

type Emulator struct {
	c *hardware.Cpu
}

func NewEmulator(romPath string) {
	cfg := config.Config{ScreenSizeX: 64, ScreenSizeY: 32} // Default CHIP8 screen size

	f, err := os.Open(romPath)
	if err != nil {
		panic(err)
	}
	fi, _ := f.Stat()

	rom := make([]byte, fi.Size())
	binary.Read(f, binary.BigEndian, &rom)
	screenBuffer := make(chan hardware.Screenbuffer)
	keyboardChan := make(chan byte)

	emulator := &Emulator{
		hardware.NewCPU(cfg, screenBuffer, keyboardChan),
	}

	emulator.c.LoadROM(rom)
	emulator.c.StartTimer()
	go emulator.c.FE()

	pixelgl.Run(func() {
		hardware.NewScreen(cfg, screenBuffer, keyboardChan)
	})
}
