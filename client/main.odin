//+build js
package main

import "base:runtime"
import "vendor:wasm/js"

HTML_Canvas_ID :: "wasm"

Context :: struct {
	initialized:      bool,
	accumulated_time: f64,
}

state: Context = {
	initialized      = false,
	accumulated_time = 1,
}

main :: proc() {}

run :: proc() {
	state.initialized = true
}

@(export)
step :: proc(delta_time: f64) -> (keep_going: bool) {
	if !state.initialized {
		return true
	}
	state.accumulated_time += delta_time

	if state.accumulated_time > 1 {
		frame(state.accumulated_time)
	}

	for state.accumulated_time > 1 {
		state.accumulated_time -= 1
	}

	return true
}

frame :: proc(delta_time: f64) {}

@(private = "file", fini)
finish :: proc() {}
