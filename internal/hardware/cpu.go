package hardware

import (
	"encoding/binary"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/trondhumbor/chip8/internal/config"
)

type Cpu struct {
	ram          []byte
	v            []byte
	i            uint16
	dt           timer
	st           timer
	pc           uint16
	sp           int8
	stack        []uint16
	screenBuffer Screenbuffer
	screenChan   chan<- Screenbuffer
	keyboardChan <-chan byte
	cfg          config.Config
}

type timer struct {
	mu    sync.Mutex
	value byte
}

func (t *timer) Dec() {
	t.mu.Lock()
	t.value--
	t.mu.Unlock()
}

func (t *timer) Set(value byte) {
	t.mu.Lock()
	t.value = value
	t.mu.Unlock()
}

func (t *timer) Get() byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.value
}

func NewCPU(cfg config.Config, screenChan chan<- Screenbuffer, keyboardChan <-chan byte) *Cpu {
	cpu := &Cpu{
		make([]byte, 4096),
		make([]byte, 16),
		0,
		timer{},
		timer{},
		0x200,
		-1,
		make([]uint16, 16),
		*NewScreenbuffer(cfg.ScreenSizeX, cfg.ScreenSizeY),
		screenChan,
		keyboardChan,
		cfg,
	}

	systemSprites := []byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
	copy(cpu.ram, systemSprites)

	return cpu
}

func (c *Cpu) StartTimer() {
	tickRate := 16666666 * time.Nanosecond // 60Hz
	ticker := time.NewTicker(tickRate)
	go func() {
		for {
			select {
			case <-ticker.C:
				delayTimer := c.dt.Get()
				soundTimer := c.st.Get()

				if delayTimer > 0 {
					c.dt.Dec()
				}

				if soundTimer > 0 {
					c.st.Dec()
				}
			}
		}
	}()
}

func (c *Cpu) LoadROM(rom []byte) {
	copy(c.ram[0x200:], rom)
}

func (c *Cpu) op00E0(inst uint16) time.Duration {
	c.screenBuffer = *NewScreenbuffer(c.cfg.ScreenSizeX, c.cfg.ScreenSizeX)

	return time.Duration(109) * time.Microsecond
}

func (c *Cpu) op00EE(inst uint16) time.Duration {
	// RET
	c.pc = c.stack[c.sp]
	c.sp--

	return time.Duration(105) * time.Microsecond
}

func (c *Cpu) op1nnn(inst uint16) time.Duration {
	// JP addr
	c.pc = inst & 0x0FFF

	return time.Duration(105) * time.Microsecond
}

func (c *Cpu) op2nnn(inst uint16) time.Duration {
	// CALL addr
	c.sp++
	c.stack[c.sp] = c.pc
	c.pc = inst & 0x0FFF

	return time.Duration(105) * time.Microsecond
}

func (c *Cpu) op3xkk(inst uint16) time.Duration {
	// SE Vx, byte
	vx := (inst >> 8) & 0x0F
	kk := inst & 0x00FF
	if uint16(c.v[vx]) == kk {
		c.pc += 2
	}

	return time.Duration(55) * time.Microsecond
}

func (c *Cpu) op4xkk(inst uint16) time.Duration {
	// SNE Vx, byte
	vx := (inst >> 8) & 0x0F
	kk := inst & 0x00FF
	if uint16(c.v[vx]) != kk {
		c.pc += 2
	}

	return time.Duration(55) * time.Microsecond
}

func (c *Cpu) op5xy0(inst uint16) time.Duration {
	// SE Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	if c.v[vx] == c.v[vy] {
		c.pc += 2
	}

	return time.Duration(73) * time.Microsecond
}

func (c *Cpu) op6xkk(inst uint16) time.Duration {
	// LD Vx, byte
	vx := (inst >> 8) & 0x0F
	c.v[vx] = byte(inst & 0x00FF)

	return time.Duration(27) * time.Microsecond
}

func (c *Cpu) op7xkk(inst uint16) time.Duration {
	// ADD Vx, byte
	vx := (inst >> 8) & 0x0F
	c.v[vx] += byte(inst & 0x00FF)

	return time.Duration(45) * time.Microsecond
}

func (c *Cpu) op8xy0(inst uint16) time.Duration {
	// LD Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	c.v[vx] = c.v[vy]

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xy1(inst uint16) time.Duration {
	// OR Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	c.v[vx] |= c.v[vy]

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xy2(inst uint16) time.Duration {
	// AND Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	c.v[vx] &= c.v[vy]

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xy3(inst uint16) time.Duration {
	// XOR Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	c.v[vx] ^= c.v[vy]

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xy4(inst uint16) time.Duration {
	// ADD Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	c.v[0xF] = 0
	if c.v[vx]+c.v[vy] > 0xFF {
		c.v[0xF] = 1
	}
	c.v[vx] = (c.v[vx] + c.v[vy]) & 0xFF

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xy5(inst uint16) time.Duration {
	// SUB Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	c.v[0xF] = 0
	if c.v[vx] > c.v[vy] {
		c.v[0xF] = 1
	}
	c.v[vx] -= c.v[vy]

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xy6(inst uint16) time.Duration {
	// SHR Vx
	vx := (inst >> 8) & 0x0F
	c.v[0xF] = c.v[vx] & 0x1
	c.v[vx] /= 2

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xy7(inst uint16) time.Duration {
	// SUBN Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	c.v[0xF] = 0
	if c.v[vy] > c.v[vx] {
		c.v[0xF] = 1
	}
	c.v[vx] = c.v[vy] - c.v[vx]

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op8xyE(inst uint16) time.Duration {
	// SHR Vx
	vx := (inst >> 8) & 0x0F
	c.v[0xF] = c.v[vx] >> 7 & 0x1
	c.v[vx] *= 2

	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) op9xy0(inst uint16) time.Duration {
	// SNE Vx, Vy
	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	if c.v[vx] != c.v[vy] {
		c.pc -= 4
	}

	return time.Duration(73) * time.Microsecond
}

func (c *Cpu) opAnnn(inst uint16) time.Duration {
	// LD, I, addr
	c.i = inst & 0x0FFF

	return time.Duration(55) * time.Microsecond
}

func (c *Cpu) opBnnn(inst uint16) time.Duration {
	// JP V0, addr
	c.pc = uint16(c.v[0]) + (inst & 0x0FFF)

	return time.Duration(105) * time.Microsecond
}

func (c *Cpu) opCxkk(inst uint16) time.Duration {
	// RND Vx, byte
	vx := (inst >> 8) & 0x0F
	kk := inst & 0xFF
	rand.Seed(time.Now().UnixNano())
	c.v[vx] = byte(rand.Intn(256)) & byte(kk)

	return time.Duration(164) * time.Microsecond
}

func (c *Cpu) opDxyn(inst uint16) time.Duration {
	// DRW Vx, Vy, nibble
	defaultSpriteWidth := 8

	vx := (inst >> 8) & 0x0F
	vy := (inst >> 4) & 0x00F
	nibble := inst & 0x0F

	spriteToPixels := func(sprite []byte) Screenbuffer {
		pixelBuffer := *NewScreenbuffer(defaultSpriteWidth, int(nibble))
		for spriteY := 0; spriteY < int(nibble); spriteY++ {
			for spriteX := defaultSpriteWidth - 1; spriteX >= 0; spriteX-- {
				pixelSet := int(sprite[spriteY] >> spriteX & 0x1)
				pixelBuffer.Put(defaultSpriteWidth-spriteX-1, spriteY, pixelSet)
			}
		}
		return pixelBuffer
	}

	sprite := []byte{}
	for j := uint16(0); j < nibble; j++ {
		sprite = append(sprite, c.ram[c.i+j])
	}

	pixels := spriteToPixels(sprite)

	c.v[0xF] = 0
	for y := 0; y < int(nibble); y++ {
		for x := 0; x < defaultSpriteWidth; x++ {

			screenPosX := (int(c.v[vx]) + x) % c.cfg.ScreenSizeX
			screenPosY := (int(c.v[vy]) + y) % c.cfg.ScreenSizeY

			currPixel := c.screenBuffer.Get(screenPosX, screenPosY)
			newPixel := pixels.Get(x, y)

			var collision bool
			// Basic XOR
			if currPixel^newPixel == 1 {
				collision = c.screenBuffer.Put(
					screenPosX,
					screenPosY,
					1,
				)
			} else {
				collision = c.screenBuffer.Put(
					screenPosX,
					screenPosY,
					0,
				)
			}
			if collision && c.v[0xF] == 0 {
				c.v[0xF] = 1
			}
		}
	}

	toBeSent := c.screenBuffer

	c.screenChan <- toBeSent

	// return  time.Duration(22734) * time.Microsecond
	return time.Duration(15000) * time.Microsecond
}

func (c *Cpu) opEx9E(inst uint16) time.Duration {
	// SKP Vx
	vx := (inst >> 8) & 0x0F
	key := c.v[vx]

	select {
	case k := <-c.keyboardChan:
		if k == key {
			c.pc += 2
		}
	case <-time.After(60 * time.Microsecond):
	}

	return time.Duration(73) * time.Microsecond
}

func (c *Cpu) opExA1(inst uint16) time.Duration {
	// SKNP Vx
	vx := (inst >> 8) & 0x0F
	key := c.v[vx]

	select {
	case k := <-c.keyboardChan:
		if k != key {
			c.pc += 2
		}
	case <-time.After(60 * time.Microsecond):
		c.pc += 2
	}

	return time.Duration(73) * time.Microsecond
}

func (c *Cpu) opFx07(inst uint16) time.Duration {
	// LD Vx, DT
	vx := (inst >> 8) & 0x0F
	c.v[vx] = c.dt.Get()
	return time.Duration(45) * time.Microsecond
}

func (c *Cpu) opFx0A(inst uint16) time.Duration {
	// LD Vx, K
	vx := (inst >> 8) & 0x0F
	select {
	case k := <-c.keyboardChan:
		c.v[vx] = k
	}

	return time.Duration(0) * time.Microsecond
}

func (c *Cpu) opFx15(inst uint16) time.Duration {
	// LD DT, Vx
	vx := (inst >> 8) & 0x0F
	c.dt.Set(c.v[vx])
	return time.Duration(45) * time.Microsecond
}

func (c *Cpu) opFx18(inst uint16) time.Duration {
	// LD ST, Vx
	vx := (inst >> 8) & 0x0F
	c.st.Set(c.v[vx])
	return time.Duration(45) * time.Microsecond
}

func (c *Cpu) opFx1E(inst uint16) time.Duration {
	// ADD I, Vx
	vx := (inst >> 8) & 0x0F
	c.i += uint16(c.v[vx])
	return time.Duration(86) * time.Microsecond
}

func (c *Cpu) opFx29(inst uint16) time.Duration {
	// LD F, Vx
	vx := (inst >> 8) & 0x0F
	digit := c.v[vx]
	c.i = uint16(digit) * 5
	return time.Duration(91) * time.Microsecond
}

func (c *Cpu) opFx33(inst uint16) time.Duration {
	// LD B, Vx
	vx := (inst >> 8) & 0x0F
	digit := c.v[vx]

	getDigitAtPos := func(number, pos int) int {
		return number / int(math.Pow10(pos)) % 10
	}

	hundreds, tens, ones := getDigitAtPos(int(digit), 2),
		getDigitAtPos(int(digit), 1),
		getDigitAtPos(int(digit), 0)

	c.ram[c.i], c.ram[c.i+1], c.ram[c.i+2] = byte(hundreds), byte(tens), byte(ones)

	//return  time.Duration(364 + (hundreds+tens+ones)*73) * time.Microsecond
	return time.Duration(927) * time.Microsecond
}

func (c *Cpu) opFx55(inst uint16) time.Duration {
	// LD [I], Vx
	vx := (inst >> 8) & 0x0F
	for j := uint16(0); j <= vx; j++ {
		c.ram[c.i+j] = c.v[j]
	}

	//return  time.Duration(64 + vx*64) * time.Microsecond
	return time.Duration(605) * time.Microsecond
}

func (c *Cpu) opFx65(inst uint16) time.Duration {
	// LD Vx, [i]
	vx := (inst >> 8) & 0x0F
	for j := uint16(0); j <= vx; j++ {
		c.v[j] = c.ram[c.i+j]
	}

	//return  time.Duration(64 + vx*64)
	return time.Duration(605) * time.Microsecond
}

func (c *Cpu) opNOOP(inst uint16) time.Duration {
	// NOOP
	return time.Duration(200) * time.Microsecond
}

func (c *Cpu) FE() {
	for {
		instruction := binary.BigEndian.Uint16(c.ram[c.pc : c.pc+2])
		highnibble := instruction >> 12
		lowbyte := instruction & 0xFF
		c.pc += 2

		startTime := time.Now()
		var opDuration time.Duration
		switch highnibble {
		case 0:
			switch lowbyte {
			case 0xE0:
				opDuration = c.op00E0(instruction)
			case 0xEE:
				opDuration = c.op00EE(instruction)
			}
		case 1:
			opDuration = c.op1nnn(instruction)
		case 2:
			opDuration = c.op2nnn(instruction)
		case 3:
			opDuration = c.op3xkk(instruction)
		case 4:
			opDuration = c.op4xkk(instruction)
		case 5:
			opDuration = c.op5xy0(instruction)
		case 6:
			opDuration = c.op6xkk(instruction)
		case 7:
			opDuration = c.op7xkk(instruction)
		case 8:
			lownibble := lowbyte & 0b0000_1111
			switch lownibble {
			case 0x0:
				opDuration = c.op8xy0(instruction)
			case 0x1:
				opDuration = c.op8xy1(instruction)
			case 0x2:
				opDuration = c.op8xy2(instruction)
			case 0x3:
				opDuration = c.op8xy3(instruction)
			case 0x4:
				opDuration = c.op8xy4(instruction)
			case 0x5:
				opDuration = c.op8xy5(instruction)
			case 0x6:
				opDuration = c.op8xy6(instruction)
			case 0x7:
				opDuration = c.op8xy7(instruction)
			case 0xE:
				opDuration = c.op8xyE(instruction)
			}
		case 9:
			opDuration = c.op9xy0(instruction)
		case 0xA:
			opDuration = c.opAnnn(instruction)
		case 0xB:
			opDuration = c.opBnnn(instruction)
		case 0xC:
			opDuration = c.opCxkk(instruction)
		case 0xD:
			opDuration = c.opDxyn(instruction)
		case 0xE:
			switch lowbyte {
			case 0x9E:
				opDuration = c.opEx9E(instruction)
			case 0xA1:
				opDuration = c.opExA1(instruction)
			}
		case 0xF:
			switch lowbyte {
			case 0x07:
				opDuration = c.opFx07(instruction)
			case 0x0A:
				opDuration = c.opFx0A(instruction)
			case 0x15:
				opDuration = c.opFx15(instruction)
			case 0x18:
				opDuration = c.opFx18(instruction)
			case 0x1E:
				opDuration = c.opFx1E(instruction)
			case 0x29:
				opDuration = c.opFx29(instruction)
			case 0x33:
				opDuration = c.opFx33(instruction)
			case 0x55:
				opDuration = c.opFx55(instruction)
			case 0x65:
				opDuration = c.opFx65(instruction)
			}
		}

		timeUsed := time.Since(startTime)
		if timeUsed < opDuration {
			time.Sleep(opDuration - timeUsed)
		}
	}
}
