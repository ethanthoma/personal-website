//+build js
package main

import "base:runtime"
import "vendor:wasm/js"
import "vendor:wgpu"

HTML_Canvas_ID :: "wasm"

Context :: struct {
	initialized:      bool,
	accumulated_time: f64,
	instance:         wgpu.Instance,
	config:           wgpu.SurfaceConfiguration,
	surface:          wgpu.Surface,
}

state: Context = {
	initialized      = false,
	accumulated_time = 1,
}

main :: proc() {
	init()

	state.instance = wgpu.CreateInstance(nil)
	if state.instance == nil {
		panic("WebGPU is not supported")
	}

	state.surface = get_surface(state.instance)
}

init :: proc() {
	assert(js.add_window_event_listener(.Resize, nil, size_callback))
}

run :: proc() {
	state.initialized = true
}

get_surface :: proc(instance: wgpu.Instance) -> wgpu.Surface {
	return wgpu.InstanceCreateSurface(
		instance,
		&wgpu.SurfaceDescriptor {
			nextInChain = &wgpu.SurfaceDescriptorFromCanvasHTMLSelector {
				sType = .SurfaceDescriptorFromCanvasHTMLSelector,
				selector = "#" + HTML_Canvas_ID,
			},
		},
	)
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

frame :: proc(delta_time: f64) {

}

@(private = "file", fini)
finish :: proc() {
	js.remove_window_event_listener(.Resize, nil, size_callback)

	wgpu.InstanceRelease(state.instance)
}

get_render_bounds :: proc() -> (width, height: u32) {
	rect := js.get_bounding_client_rect("body")
	return u32(rect.width), u32(rect.height)
}

@(private = "file")
size_callback :: proc(e: js.Event) {
	state.config.width, state.config.height = get_render_bounds()
	wgpu.SurfaceConfigure(state.surface, &state.config)
}
