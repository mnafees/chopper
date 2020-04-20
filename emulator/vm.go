package emulator

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	totalMemory    = 0x1000
	pcStartLoc     = 0x200
	maxProgramSize = totalMemory - pcStartLoc
	timerHz        = float64(1000.0 / 60)
)

// C8VM is an emulated CHIP-8 VM
type C8VM struct {
	opcode     uint16             // 16-bit opcode of the current instruction
	regV       [16]uint8          // 16 general purpose 8-bit registers
	regI       uint16             // 16-bit register that is generally used to store memory addresses
	delayTimer uint8              // Delay timer
	soundTimer uint8              // Sound timer
	pc         uint16             // Program counter
	sp         uint8              // Stack pointer
	stack      [16]uint16         // A stack of 16 16-bit values
	memory     [totalMemory]uint8 // 4 KB global memory
	io         *IO                // I/O layer
	prevTime   time.Time          // Just a counter to keep the previously logged time
}

var fontset = []uint8{
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

// NewC8VM creates a new instance of an emulated CHIP-8 VM
func NewC8VM(ioLayer *IO) *C8VM {
	vm := &C8VM{
		pc:       pcStartLoc,
		io:       ioLayer,
		prevTime: time.Now(),
	}
	if copy(vm.memory[:], fontset[:]) != len(fontset) {
		fmt.Println("Error copying fontset data to memory")
		os.Exit(1)
	}
	return vm
}

// LoadProgram loads a given CHIP-8 program into the VM's memory
func (vm *C8VM) LoadProgram(filename string) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Printf("Error loading program: %v\n", err)
		os.Exit(1)
	}
	dataSize := len(data)
	if dataSize > maxProgramSize {
		fmt.Println("Program size exceeds the maximum size")
		os.Exit(1)
	}
	if copy(vm.memory[pcStartLoc:], data) != dataSize {
		fmt.Println("Error copying program data into VM's memory")
		os.Exit(1)
	}
}

// Loop is the main application loop
func (vm *C8VM) Loop() {
	running := true
	for running {
		vm.readNextInstruction()

		if vm.io.clearFlag {
			vm.io.clearScreen()
		}

		if vm.io.drawFlag {
			vm.io.drawSprite()
		}

		if float64(time.Since(vm.prevTime).Milliseconds()) >= timerHz {
			if vm.delayTimer > 0 {
				vm.delayTimer--
			}
			if vm.soundTimer > 0 {
				vm.soundTimer--
			}
			vm.prevTime = time.Now()
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				keycode := t.Keysym.Scancode
				switch t.GetType() {
				case sdl.KEYDOWN:
					vm.io.setKeymask(keycode)
					break
				case sdl.KEYUP:
					vm.io.unsetKeymask(keycode)
				}
				break
			case *sdl.QuitEvent:
				running = false
				break
			}
		}
	}
}

func (vm *C8VM) unknownOpcode() {
	fmt.Printf("Unknown opcode: %04X\n", vm.opcode)
	os.Exit(1)
}

func (vm *C8VM) initSprite(x uint8, y uint8, n uint8) {
	if n > 15 || n == 0 {
		vm.unknownOpcode()
	}
	vm.regV[0xF] = 0
	for byteIdx := uint8(0); byteIdx < n; byteIdx++ {
		spriteByte := vm.memory[vm.regI+uint16(byteIdx)]
		for bitIdx := uint8(0); bitIdx < 8; bitIdx++ {
			bit := (spriteByte >> bitIdx) & 0x1
			px := &vm.io.pixels[(x+(7-bitIdx))%screenWidth][(y+byteIdx)%screenHeight]
			if bit == 1 && *px == 1 {
				vm.regV[0xF] = 1
			}
			*px ^= bit
		}
	}
}

func (vm *C8VM) readNextInstruction() {
	vm.opcode = uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc+1]) // 16-bit instruction opcode
	x := uint8((vm.opcode >> 8) & 0x000F)                                // the lower 4 bits of the high byte of the instruction
	y := uint8((vm.opcode >> 4) & 0x000F)                                // the upper 4 bits of the low byte of the instruction
	n := uint8(vm.opcode & 0x000F)                                       // the lowest 4 bits of the instruction
	kk := uint8(vm.opcode & 0x00FF)                                      // the lowest 8 bits of the instruction
	nnn := uint16(vm.opcode & 0x0FFF)                                    // the lowest 12 bits of the instruction

	switch vm.opcode & 0xF000 { // Compare against the first 4 bits of the instruction only
	case 0x0000:
		switch kk {
		case 0xE0: // CLS
			vm.io.clearFlag = true
			vm.pc += 2
			break
		case 0xEE: // RET
			vm.sp--
			vm.pc = vm.stack[vm.sp] + 2
			break
		default:
			vm.unknownOpcode()
		}
		break
	case 0x1000: // JP nnn
		vm.pc = nnn
		break
	case 0x2000: // CALL nnn
		vm.stack[vm.sp] = vm.pc
		vm.sp++
		vm.pc = nnn
		break
	case 0x3000: // SE Vx, kk
		if vm.regV[x] == kk {
			vm.pc += 2
		}
		vm.pc += 2
		break
	case 0x4000: // SNE Vx, kk
		if vm.regV[x] != kk {
			vm.pc += 2
		}
		vm.pc += 2
		break
	case 0x5000:
		switch n {
		case 0x0: // SE Vx, Vy
			if vm.regV[x] == vm.regV[y] {
				vm.pc += 2
			}
			vm.pc += 2
			break
		default:
			vm.unknownOpcode()
		}
		break
	case 0x6000: // LD Vx, kk
		vm.regV[x] = kk
		vm.pc += 2
		break
	case 0x7000: // ADD Vx, kk
		vm.regV[x] += kk
		vm.pc += 2
		break
	case 0x8000:
		switch n {
		case 0x0: // LD Vx, Vy
			vm.regV[x] = vm.regV[y]
			break
		case 0x1: // OR Vx, Vy
			vm.regV[x] |= vm.regV[y]
			break
		case 0x2: // AND Vx, Vy
			vm.regV[x] &= vm.regV[y]
			break
		case 0x3: // XOR Vx, Vy
			vm.regV[x] ^= vm.regV[y]
			break
		case 0x4: // ADD Vx, Vy
			temp := uint16(vm.regV[x]) + uint16(vm.regV[y])
			if temp > 255 {
				vm.regV[0xF] = 1
			} else {
				vm.regV[0xF] = 0
			}
			vm.regV[x] = uint8(temp & 0x0000FFFF)
			break
		case 0x5: // SUB Vx, Vy
			if vm.regV[x] > vm.regV[y] {
				vm.regV[0xF] = 1
			} else {
				vm.regV[0xF] = 0
			}
			vm.regV[x] -= vm.regV[y]
			break
		case 0x6: // SHR Vx {, Vy}
			if vm.regV[x]&0x01 == 1 {
				vm.regV[0xF] = 1
			} else {
				vm.regV[0xF] = 0
			}
			vm.regV[x] /= 2
			break
		case 0x7: // SUBN Vx, Vy
			if vm.regV[y] > vm.regV[x] {
				vm.regV[0xF] = 1
			} else {
				vm.regV[0xF] = 0
			}
			vm.regV[x] = vm.regV[y] - vm.regV[x]
			break
		case 0xE: // SHL Vx {, Vy}
			if vm.regV[x]&0x80 == 0x80 {
				vm.regV[0xF] = 1
			} else {
				vm.regV[0xF] = 0
			}
			vm.regV[x] *= 2
			break
		default:
			vm.unknownOpcode()
		}
		vm.pc += 2
		break
	case 0x9000:
		switch n {
		case 0x0: // SNE Vx, Vy
			if vm.regV[x] != vm.regV[y] {
				vm.pc += 2
			}
			vm.pc += 2
			break
		default:
			vm.unknownOpcode()
		}
		break
	case 0xA000: // LD I, nnn
		vm.regI = nnn
		vm.pc += 2
		break
	case 0xB000: // JP V0, nnn
		vm.pc = nnn + uint16(vm.regV[0])
		break
	case 0xC000: // RND Vx, kk
		vm.regV[x] = uint8(rand.Intn(256)) & kk
		vm.pc += 2
		break
	case 0xD000: // DRW Vx, Vy, n
		vm.initSprite(vm.regV[x], vm.regV[y], n)
		vm.pc += 2
		vm.io.drawFlag = true
		break
	case 0xE000:
		switch kk {
		case 0x9E: // SKP Vx
			if vm.io.issetKeymask(vm.regV[x]) {
				vm.pc += 2
			}
			break
		case 0xA1: // SKNP Vx
			if !vm.io.issetKeymask(vm.regV[x]) {
				vm.pc += 2
			}
			break
		default:
			vm.unknownOpcode()
		}
		vm.pc += 2
		break
	case 0xF000:
		switch kk {
		case 0x07: // LD Vx, DT
			vm.regV[x] = vm.delayTimer
			break
		case 0x0A: // LD Vx, K
			loop := true
			for loop {
				for i := uint8(0x0); i <= 0xF; i++ {
					if vm.io.issetKeymask(vm.regV[x]) {
						vm.regV[x] = i
						loop = false
						break
					}
				}
			}
			break
		case 0x15: // LD DT, Vx
			vm.delayTimer = vm.regV[x]
			break
		case 0x18: // LD ST, Vx
			vm.soundTimer = vm.regV[x]
			break
		case 0x1E: // ADD I, Vx
			vm.regI += uint16(vm.regV[x])
			break
		case 0x29: // LD F, Vx
			vm.regI = uint16(5 * vm.regV[x])
			break
		case 0x33: // LD B, Vx
			vm.memory[vm.regI] = (vm.regV[x] / 100) % 10
			vm.memory[vm.regI+1] = (vm.regV[x] / 10) % 10
			vm.memory[vm.regI+2] = vm.regV[x] % 10
			break
		case 0x55: // LD [I], Vx
			for i := uint16(0); i <= uint16(x); i++ {
				vm.memory[vm.regI+i] = vm.regV[i]
			}
			break
		case 0x65: // LD Vx, [I]
			for i := uint16(0); i <= uint16(x); i++ {
				vm.regV[i] = vm.memory[vm.regI+i]
			}
			break
		default:
			vm.unknownOpcode()
		}
		vm.pc += 2
		break
	default:
		vm.unknownOpcode()
	}
}
