package emulator

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	screenWidth  = 64
	screenHeight = 32
	pixelSize    = 20

	screenColor = 0x1A237E
	spriteColor = 0x9FA8DA
)

// IO is the input/output abstraction layer for the VM
type IO struct {
	window  *sdl.Window
	surface *sdl.Surface

	// 64 px x 32 px display
	pixels [screenWidth][screenHeight]uint8
	// A 16-bit integer to hold the current key values in the form of individual bits.
	// So when 0 is pushed in the keypad, the 0'th bit will be set and so on.
	key int16

	clearFlag bool
	drawFlag  bool
}

// NewIO returns a new IO instance to be provided to the VM
func NewIO() *IO {
	return &IO{}
}

// SetupWindow initialises and sets up the main SDL window
func (io *IO) SetupWindow(title string) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Printf("Error initialising SDL: %v", err)
		os.Exit(1)
	}

	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		screenWidth*pixelSize, screenHeight*pixelSize, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Printf("Error creating window: %v", err)
		os.Exit(1)
	}
	io.window = window
	io.surface, err = window.GetSurface()
	if err != nil {
		fmt.Printf("Error getting window surface: %v\n", err)
		os.Exit(1)
	}
	io.surface.FillRect(nil, screenColor)
}

// Destroy should be called before quitting the application
func (io *IO) Destroy() {
	io.window.Destroy()
	sdl.Quit()
}

// Clear the current application screen
func (io *IO) clearScreen() {
	for w := 0; w < screenWidth; w++ {
		for h := 0; h < screenHeight; h++ {
			io.pixels[w][h] = 0
		}
	}
	io.surface.FillRect(nil, screenColor)
	io.window.UpdateSurface()
	io.clearFlag = false
}

// Draws a sprite
func (io *IO) drawSprite() {
	io.surface.FillRect(nil, screenColor)
	for w := int32(0); w < screenWidth; w++ {
		for h := int32(0); h < screenHeight; h++ {
			if io.pixels[w][h] == 1 {
				rect := &sdl.Rect{w * pixelSize, h * pixelSize, pixelSize, pixelSize}
				io.surface.FillRect(rect, spriteColor)
			}
		}
	}
	io.window.UpdateSurface()
	io.drawFlag = false
}

// Maps keys from a QWERTY keyboard to the keypad used by CHIP-8
// Below we have a mapping QWERTY keyboard to the CHIP-8 keypad
// +--------+--------+--------+--------+
// | 1 -> 1 | 2 -> 2 | 3 -> 3 | 4 -> C |
// +--------+--------+--------+--------+
// | Q -> 4 | W -> 5 | E -> 6 | R -> D |
// +--------+--------+--------+--------+
// | A -> 7 | S -> 8 | D -> 9 | F -> E |
// +--------+--------+--------+--------+
// | Z -> A | X -> 0 | C -> B | V -> F |
// +--------+--------+--------+--------+
func (io *IO) keymap(code sdl.Scancode) int8 {
	switch code {
	case sdl.SCANCODE_1:
		return 0x1
	case sdl.SCANCODE_2:
		return 0x2
	case sdl.SCANCODE_3:
		return 0x3
	case sdl.SCANCODE_4:
		return 0xC
	case sdl.SCANCODE_Q:
		return 0x4
	case sdl.SCANCODE_W:
		return 0x5
	case sdl.SCANCODE_E:
		return 0x6
	case sdl.SCANCODE_R:
		return 0xD
	case sdl.SCANCODE_A:
		return 0x7
	case sdl.SCANCODE_S:
		return 0x8
	case sdl.SCANCODE_D:
		return 0x9
	case sdl.SCANCODE_F:
		return 0xE
	case sdl.SCANCODE_Z:
		return 0xA
	case sdl.SCANCODE_X:
		return 0x0
	case sdl.SCANCODE_C:
		return 0xB
	case sdl.SCANCODE_V:
		return 0xF
	default:
		return -1
	}
}

func (io *IO) setKeymask(keycode sdl.Scancode) {
	code := io.keymap(keycode)
	if code != -1 {
		io.key |= (1 << code)
	}
}

func (io *IO) unsetKeymask(keycode sdl.Scancode) {
	code := io.keymap(keycode)
	if code != -1 {
		io.key ^= (1 << code)
	}
}

func (io *IO) issetKeymask(keycode uint8) bool {
	mask := int16(1 << keycode)
	return io.key&mask == mask
}
