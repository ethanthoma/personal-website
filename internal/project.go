package internal

import "time"

type Project struct {
	Title, Url, Description string
	Lang                    []string
	Date                    time.Time
	WIP                     bool `default:"false"`
}

var Projects = []Project{
	{
		Date:  time.Date(2025, 3, 5, 0, 0, 0, 0, time.Local),
		Title: "Glum: Mobile Game Engine in Gleam",
		Url:   "https://github.com/ethanthoma/glum",
		Description: `Elm-based mobile-game engine written in Gleam.

        Created JS FFI bindings to nativescript core libraries. This allows compiling Gleam to mobile.

        It has two render functions: UI and game objects. The UI render function includes a simple 2D layout engine, primarily written in Gleam. The game object render function has more flexibility and maps closer to Three.js primitives.
        `,
		Lang: []string{"Gleam", "JavaScript"},
		WIP:  true,
	},
	{
		Date:  time.Date(2025, 2, 11, 0, 0, 0, 0, time.Local),
		Title: "Nix builder to wrap Erlang-target Gleam code",
		Url:   "https://github.com/ethanthoma/nix-gleam-burrito",
		Description: `A Nix flake that packages Gleam projects into a single, standalone executable.

        The builder leverages the Burrito Elixir library, which lets users "compile" mix projects into a self-contained binary.

        To convert a Gleam project into a mix project, the builder injects a mix config and boilerplate application code. Then, the builder installs a plugin for compiling Gleam code.
		`,
		Lang: []string{"Nix"},
	},
	{
		Date:  time.Date(2025, 1, 25, 0, 0, 0, 0, time.Local),
		Title: "Canvas Group Quiz creation CLI in Gleam",
		Url:   "https://github.com/STASER-Lab/cgq",
		Description: `A command-line tool for the Canvas LMS that automates the creation of group-specific quizzes.

		Canvas doesn't allow educators to easily generate unique quizzes for different student groups, this tool does!`,
		Lang: []string{"Gleam"},
		WIP:  true,
	},
	{
		Date:  time.Date(2024, 12, 10, 0, 0, 0, 0, time.Local),
		Title: "Zig native WebGPU Voxel Render",
		Url:   "https://github.com/ethanthoma/graphics",
		Description: `An exploration into voxel 3D graphics programming using Zig and WebGPU.

        WebGPU is a modern graphics API. This project relies on bindings on the C headers for WebGPU.

        I (ab)use Zig's comptime to generate structs and other objects for my rendering logic.

        The rendering is mostly efficient, using compacted data structures. Chunking for voxels is also implemented.

        The main limitation on performance is that the application is single threaded.
		`,
		Lang: []string{"Zig"},
	},
	{
		Date:  time.Date(2024, 9, 11, 0, 0, 0, 0, time.Local),
		Title: "Interaction Nets in Odin",
		Url:   "https://github.com/ethanthoma/interaction-net",
		Description: `A runtime for Interaction Nets, a graph-based model of computation, written in Odin.

        Interaction nets are a (not-so) novel model of computation. It allows for high parallelism with ease.

        Most implementations of this model are pretty slow (save for HVM).

        I wanted to learn the Odin langauge (a systems language like C) and decided to build a HVM-based runtime.

		This project includes a basic tokenizer, parser, semantic checks, and runtime.`,
		Lang: []string{"Odin"},
	},
	{
		Date:  time.Date(2024, 7, 8, 0, 0, 0, 0, time.Local),
		Title: "Zensor: a Zig tensor library",
		Url:   "https://github.com/ethanthoma/zensor",
		Description: `A tensor library for Zig that prioritizes correctness and compile-time safety.

        Comes with compile-time known tensor shapes and types, eliminating a class of runtime errors tensors normally have.

		Tensor operations are converted into an AST. They are only executed when needed (i.e. lazy).

        When executed, the AST is converted into IR. This IR gets JITed into x86 assembly, and then executed.

        This library is basically a fully custom compiler for tensors.
        `,
		Lang: []string{"Zig"},
		WIP:  true,
	},
}
