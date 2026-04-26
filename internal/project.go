package internal

import "time"

type Project struct {
	Title, Url string
	Date       time.Time
	WIP        bool
}

var Projects = []Project{
	{
		Date:  time.Date(2025, 11, 2, 0, 0, 0, 0, time.Local),
		Title: "Window: cross-platform windowing for Zig",
		Url:   "https://github.com/ethanthoma/window",
	},
	{
		Date:  time.Date(2025, 7, 11, 0, 0, 0, 0, time.Local),
		Title: "Glum: Mobile Game Engine in Gleam",
		Url:   "https://github.com/ethanthoma/glum",
		WIP:   true,
	},
	{
		Date:  time.Date(2025, 2, 11, 0, 0, 0, 0, time.Local),
		Title: "Nix builder to wrap Erlang-target Gleam code",
		Url:   "https://github.com/ethanthoma/nix-gleam-burrito",
	},
	{
		Date:  time.Date(2025, 1, 28, 0, 0, 0, 0, time.Local),
		Title: "Trellis: pretty printing tabular data in Gleam",
		Url:   "https://github.com/ethanthoma/trellis",
	},
	{
		Date:  time.Date(2025, 1, 25, 0, 0, 0, 0, time.Local),
		Title: "Canvas Group Quiz creation CLI in Gleam",
		Url:   "https://github.com/STASER-Lab/cgq",
		WIP:   true,
	},
	{
		Date:  time.Date(2024, 12, 10, 0, 0, 0, 0, time.Local),
		Title: "Zig native WebGPU Voxel Render",
		Url:   "https://github.com/ethanthoma/graphics",
	},
	{
		Date:  time.Date(2024, 9, 11, 0, 0, 0, 0, time.Local),
		Title: "Interaction Nets in Odin",
		Url:   "https://github.com/ethanthoma/interaction-net",
	},
	{
		Date:  time.Date(2024, 7, 8, 0, 0, 0, 0, time.Local),
		Title: "Zensor: a Zig tensor library",
		Url:   "https://github.com/ethanthoma/zensor",
		WIP:   true,
	},
}
