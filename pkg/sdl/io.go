package sdl

import (
	"fmt"
	"os"
	"time"

	"github.com/mnafees/chopper/internal"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	pixelSize = 20

	screenColor = 0x1A237E
	spriteColor = 0x9FA8DA
)

// IO is the input/output abstraction layer for the VM
type IO struct {
	window  *sdl.Window
	surface *sdl.Surface

	vm *internal.C8VM
}

// NewIO returns a new I/O instance for the SDL frontend
func NewIO(vm *internal.C8VM) *IO {
	return &IO{
		vm: vm,
	}
}

// SetupWindow initialises and sets up the main SDL window
func (io *IO) SetupWindow(title string) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Printf("Error initialising SDL: %v", err)
		os.Exit(1)
	}

	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		internal.ScreenWidth*pixelSize, internal.ScreenHeight*pixelSize, sdl.WINDOW_SHOWN)
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

// Loop is the main application loop
func (io *IO) Loop() {
	running := true
	for running {
		err := io.vm.ReadNextInstruction()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if io.vm.IsClearFlagSet() {
			io.clearScreen()
		}

		if io.vm.IsDrawFlagSet() {
			io.draw()
		}

		if float64(time.Since(io.vm.PrevTime()).Milliseconds()) >= internal.TimerFrequency {
			if io.vm.DelayTimer() > 0 {
				io.vm.DecrementDelayTimer()
			}
			if io.vm.SoundTimer() > 0 {
				io.vm.DecrementSoundTimer()
			}
			io.vm.UpdatePrevTime()
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				keycode := t.Keysym.Scancode
				switch t.GetType() {
				case sdl.KEYDOWN:
					io.setKeymask(keycode)
					break
				case sdl.KEYUP:
					io.unsetKeymask(keycode)
				}
				break
			case *sdl.QuitEvent:
				running = false
				break
			}
		}
	}
}

// Clear the current appl [internal.ScreenWidth][internal.ScreenHeight]byteication screen
func (io *IO) clearScreen() {
	io.vm.NullifyPixels()
	io.surface.FillRect(nil, screenColor)
	io.window.UpdateSurface()
	io.vm.UnsetClearFlag()
}

// Draws the current sprite configuration on screen
func (io *IO) draw() {
	io.surface.FillRect(nil, screenColor)
	pixels := io.vm.Pixels()
	for w := int32(0); w < internal.ScreenWidth; w++ {
		for h := int32(0); h < internal.ScreenHeight; h++ {
			if pixels[w][h] == 1 {
				rect := &sdl.Rect{w * pixelSize, h * pixelSize, pixelSize, pixelSize}
				io.surface.FillRect(rect, spriteColor)
			}
		}
	}
	io.window.UpdateSurface()
	io.vm.UnsetDrawFlag()
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
		io.vm.SetKeymask(uint8(code))
	}
}

func (io *IO) unsetKeymask(keycode sdl.Scancode) {
	code := io.keymap(keycode)
	if code != -1 {
		io.vm.UnsetKeymask(uint8(code))
	}
}
