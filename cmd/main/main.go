package main

import (
	"fmt"
	"os"

	"github.com/trondhumbor/chip8/internal/emulator"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: chip8 romfile")
		return
	}

	emulator.NewEmulator(os.Args[1])
}
