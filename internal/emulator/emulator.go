package emulator

import (
	"encoding/binary"
	"os"

	"github.com/faiface/pixel/pixelgl"
	"github.com/trondhumbor/chip8/internal/config"
	"github.com/trondhumbor/chip8/internal/hardware"
)

type Emulator struct {
	cpu *hardware.Cpu
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
	screenChan := make(chan hardware.Screenbuffer)
	keyboardChan := make(chan byte)

	emulator := Emulator{
		hardware.NewCPU(cfg, screenChan, keyboardChan),
	}

	emulator.cpu.LoadROM(rom)
	emulator.cpu.StartTimer()
	go emulator.cpu.FE()

	pixelgl.Run(func() {
		hardware.NewScreen(cfg, screenChan, keyboardChan)
	})
}
