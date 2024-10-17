{
  pname,
  version,
  pkgs,
  fetchurl,
  buildGoApplication,
  odin,
}:
let
  htmxVersion = "2.0.2";
  htmx = fetchurl {
    url = "https://github.com/bigskysoftware/htmx/releases/download/v${htmxVersion}/htmx.min.js";
    hash = "sha256-4XRtl1nsDUPFwoRFIzOjELtf1yheusSy3Jv0TXK1qIc=";
  };

  client = pkgs.callPackage ./client.nix { inherit odin; };
in
buildGoApplication {
  inherit pname version;
  src = ../.;
  pwd = ../.;
  modules = ../gomod2nix.toml;
  subPackages = [ "cmd/${pname}" ];
  nativeBuildInputs = [ pkgs.lightningcss ];
  postInstall = ''
    static=$out/cmd/webserver/public

    mkdir -p $static

    rsync -a ./cmd/webserver/public $out/cmd/webserver --exclude styles --exclude js --exclude wasm --exclude='*.css'

    mkdir -p $static/js
    cp ${htmx} $static/js/htmx.min.js
    cp -r ${client.out}/js/* $static/js

    mkdir -p $static/wasm
    cp -r ${client.out}/wasm/* $static/wasm

    lightningcss --minify --bundle ./cmd/webserver/main.css -t "> .5% or last 2 versions" -o $out/cmd/webserver/main.css

    find ./cmd/${pname} -name "*.tmpl" -exec sh -c '
      for file do
        dest="$out/cmd/${pname}/''${file#./cmd/${pname}/}"
        echo $(dirname "$dest")
        mkdir -p "$(dirname "$dest")"
        cp "$file" "$dest"
      done
    ' sh {} +
  '';
}
