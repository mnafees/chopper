package internal

// Follows the CHIP-8 technical reference found at http://devernay.free.fr/hacks/chip8/C8TECH10.HTM

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
)

// CHIP-8 VM constants
const (
	totalMemory    = 0x1000
	pcStartAddr    = 0x200
	maxProgramSize = totalMemory - pcStartAddr

	TimerFrequency = float64(1000.0 / 60)
	ScreenWidth    = 64
	ScreenHeight   = 32
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

	prevTime time.Time // Just a counter to keep the previously logged time

	clearFlag bool // Clear screen flag
	drawFlag  bool // Draw sprite flag

	// A 16-bit integer to hold the current key values in the form of individual bits.
	// So when 0 is pushed in the keypad, the 0'th bit will be set and so on.
	key int16

	// 64 px x 32 px display
	pixels [ScreenWidth][ScreenHeight]uint8
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
func NewC8VM() (*C8VM, error) {
	vm := &C8VM{
		pc:       pcStartAddr,
		prevTime: time.Now(),
	}
	if copy(vm.memory[:], fontset[:]) != len(fontset) {
		return nil, errors.New("Error copying fontset data to memory")
	}
	return vm, nil
}

// LoadProgram loads a given CHIP-8 program into the VM's memory
func (vm *C8VM) LoadProgram(filename string) error {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return fmt.Errorf("Error loading program: %v", err)
	}
	size := len(data)
	if size > maxProgramSize {
		return errors.New("Program size exceeds the maximum size")
	}
	if copy(vm.memory[pcStartAddr:], data) != size {
		return errors.New("Error copying program data into VM's memory")
	}
	return nil
}

func (vm *C8VM) unknownOpcode() error {
	return fmt.Errorf("Unknown opcode: %04X", vm.opcode)
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
			px := &vm.pixels[(x+(7-bitIdx))%ScreenWidth][(y+byteIdx)%ScreenHeight]
			if bit == 1 && *px == 1 {
				vm.regV[0xF] = 1
			}
			*px ^= bit
		}
	}
}

// NullifyPixels resets all pixels to a value of 0
func (vm *C8VM) NullifyPixels() {
	for w := 0; w < ScreenWidth; w++ {
		for h := 0; h < ScreenHeight; h++ {
			vm.pixels[w][h] = 0
		}
	}
}

// Pixels returns the pixels 2d slice
func (vm *C8VM) Pixels() [ScreenWidth][ScreenHeight]byte {
	return vm.pixels
}

// ReadNextInstruction reads the next instruction to execute
func (vm *C8VM) ReadNextInstruction() error {
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
			vm.clearFlag = true
			vm.pc += 2
			break
		case 0xEE: // RET
			vm.sp--
			vm.pc = vm.stack[vm.sp] + 2
			break
		default:
			return vm.unknownOpcode()
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
			return vm.unknownOpcode()
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
			return vm.unknownOpcode()
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
			return vm.unknownOpcode()
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
		vm.drawFlag = true
		break
	case 0xE000:
		switch kk {
		case 0x9E: // SKP Vx
			if vm.issetKeymask(vm.regV[x]) {
				vm.pc += 2
			}
			break
		case 0xA1: // SKNP Vx
			if !vm.issetKeymask(vm.regV[x]) {
				vm.pc += 2
			}
			break
		default:
			return vm.unknownOpcode()
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
					if vm.issetKeymask(vm.regV[x]) {
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
			return vm.unknownOpcode()
		}
		vm.pc += 2
		break
	default:
		return vm.unknownOpcode()
	}
	return nil
}

// SetKeymask sets the respective bit in the key
func (vm *C8VM) SetKeymask(code uint8) {
	vm.key |= (1 << code)
}

// UnsetKeymask unsets the respective bit in the key
func (vm *C8VM) UnsetKeymask(code uint8) {
	vm.key ^= (1 << code)
}

// PrevTime returns the prviously logged time
func (vm *C8VM) PrevTime() time.Time {
	return vm.prevTime
}

// IsClearFlagSet returns whether the clear flag is set
func (vm *C8VM) IsClearFlagSet() bool {
	return vm.clearFlag
}

// UnsetClearFlag unsets the clear flag
func (vm *C8VM) UnsetClearFlag() {
	vm.clearFlag = false
}

// IsDrawFlagSet returns whether the draw flag is set
func (vm *C8VM) IsDrawFlagSet() bool {
	return vm.drawFlag
}

// UnsetDrawFlag unsets the draw flag
func (vm *C8VM) UnsetDrawFlag() {
	vm.drawFlag = false
}

// DelayTimer returns the value of DT
func (vm *C8VM) DelayTimer() uint8 {
	return vm.delayTimer
}

// DecrementDelayTimer decrements the value of DT
func (vm *C8VM) DecrementDelayTimer() {
	if vm.delayTimer > 0 {
		vm.delayTimer--
	}
}

// SoundTimer returns the value of ST
func (vm *C8VM) SoundTimer() uint8 {
	return vm.soundTimer
}

// DecrementSoundTimer decrements the value of ST
func (vm *C8VM) DecrementSoundTimer() {
	if vm.soundTimer > 0 {
		vm.soundTimer--
	}
}

// UpdatePrevTime updates the previously logged time to the current time
func (vm *C8VM) UpdatePrevTime() {
	vm.prevTime = time.Now()
}

func (vm *C8VM) issetKeymask(code uint8) bool {
	mask := int16(1 << code)
	return vm.key&mask == mask
}
