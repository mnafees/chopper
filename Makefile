sdl:
	@go build -o chopper ./cmd/sdl

clean:
	@rm chopper

.PHONY: clean
