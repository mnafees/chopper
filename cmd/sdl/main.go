package main

import (
	"fmt"
	"os"

	"github.com/mnafees/chopper/internal"
	"github.com/mnafees/chopper/pkg/sdl"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: chopper <CHIP-8 program>")
		os.Exit(1)
	}

	vm, err := internal.NewC8VM()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = vm.LoadProgram(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	io := sdl.NewIO(vm)
	defer io.Destroy()
	io.SetupWindow("Chopper | CHIP-8 Emulator")
	io.Loop()
}
