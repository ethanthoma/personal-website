package main

import "vendor:wasm/js"

import "core:fmt"
import "core:strings"
import "core:time"

Context :: struct {
	accumulated_time: f64,
}

ctx: Context = {1}

@(export)
step :: proc(delta_time: f64) -> (keep_going: bool) {
	ctx.accumulated_time += delta_time

	if ctx.accumulated_time > 1 {
	}

	for ctx.accumulated_time > 1 {
		ctx.accumulated_time -= 1
	}

	return true
}
