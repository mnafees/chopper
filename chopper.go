package main

// Follows the CHIP-8 technical reference found at http://devernay.free.fr/hacks/chip8/C8TECH10.HTM

import (
	"fmt"
	"os"

	"github.com/mnafees/chopper/emulator"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: chopper <CHIP-8 program>")
		os.Exit(1)
	}

	io := emulator.NewIO()
	vm := emulator.NewC8VM(io)
	vm.LoadProgram(os.Args[1])

	io.SetupWindow("Chopper | CHIP-8 Emulator")
	defer io.Destroy()

	vm.Loop()
}
