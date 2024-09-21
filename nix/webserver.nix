{ pname
, version
, pkgs
, fetchurl
, buildGoApplication
}:
let
    htmxVersion = "2.0.2";
    htmx = fetchurl {
        url = "https://github.com/bigskysoftware/htmx/releases/download/v${htmxVersion}/htmx.min.js";
        hash = "sha256-4XRtl1nsDUPFwoRFIzOjELtf1yheusSy3Jv0TXK1qIc=";
    };
in
    buildGoApplication {
        inherit pname version;
        src = ../.;
        pwd = ../.;
        modules = ../gomod2nix.toml;
        subPackages = [ "cmd/${pname}" ];
        nativeBuildInputs = [
            pkgs.lightningcss
        ];
        postInstall = ''
        mkdir -p $out/static
        rsync -a ./static $out --exclude styles js

        mkdir -p $out/static/js
        cp ${htmx} $out/static/js/htmx.min.js

        mkdir -p $out/static/styles
        lightningcss --bundle ./static/styles/main.css -t "> .5% or last 2 versions" -o $out/static/styles/main.css

        mkdir -p $out/cmd/${pname}
        cp -rf ./cmd/${pname}/pages $out/cmd/${pname}/pages
        cp -rf ./cmd/${pname}/components $out/cmd/${pname}/components
        '';
    }
