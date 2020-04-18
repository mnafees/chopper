# Chopper

A CHIP-8 Emulator written in Go with SDL

## Dependencies

- Go 1.11+
- libsdl 2

## Build

```
$ make        # Builds the chopper executable
$ make clean  # Removes the chopper executable
```

## Usage

```
$ ./chopper <path to a CHIP-8 ROM>
```

## Test ROMs

I have used the [ROMs available here](https://github.com/dmatlack/chip8/tree/master/roms) to test the emulator. Please feel free to open an issue if you find something broken.

## License

```
MIT License

Copyright (c) 2020 Mohammed Nafees

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
