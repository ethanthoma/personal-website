{ pkgs
, odin
,
}:
pkgs.stdenv.mkDerivation
rec {
  pname = "client";
  version = "0.1";
  src = ../client;

  nativeBuildInputs = [ odin ];

  buildPhase = ''
    mkdir -p $out/wasm
    odin build $src -show-timings -out:$out/wasm/${pname}.wasm -no-bounds-check -o:speed -target:js_wasm32

    mkdir -p $out/js
    cp ${odin.src}/vendor/wasm/js/runtime.js $out/js

    touch $out/bin
  '';

  doCheck = true;

  checkPhase = ''
    odin test $src
  '';
}
