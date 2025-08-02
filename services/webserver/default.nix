{
  fetchurl,
  makeWrapper,
  buildGoApplication,
  tailwindcss,
  templpkgs,
}:
let
  htmxVersion = "2.0.3";
  htmx = fetchurl {
    url = "https://github.com/bigskysoftware/htmx/releases/download/v${htmxVersion}/htmx.min.js";
    hash = "sha256-SRlVzRgQdH19e5zLk2QAr7dg4G0l1T5FcrZLZWOyeE4=";
  };

  preloadVersion = "2.1.1";
  preload = fetchurl {
    url = "https://cdn.jsdelivr.net/npm/htmx-ext-preload@${preloadVersion}/dist/preload.min.js";
    hash = "sha256-E17eAiDgdtSal+NsO8D/QOlyCcRtW7dNzf1dhObsDCk=";
  };
in
buildGoApplication rec {
  pname = "webserver";
  version = "0.1";
  src = ../../.;
  pwd = ./.;
  modules = ../../gomod2nix.toml;
  subPackages = [ "services/${pname}" ];
  nativeBuildInputs = [
    makeWrapper
    tailwindcss
  ];
  preBuild = ''
    ${templpkgs}/bin/templ generate .
  '';
  postInstall = ''
    public=./services/${pname}/public
    static=$out/public

    mkdir -p $static

    tailwindcss -i $public/main.css -o $static/main.css --minify

    rsync -a $public $out --exclude js --exclude='*.css'

    mkdir -p $static/js
    cp ${htmx} $static/js/htmx.min.js
    cp ${preload} $static/js/preload.min.js

    mv $out/bin/${pname} $out/bin/.${pname}-unwrapped
    makeWrapper $out/bin/.${pname}-unwrapped $out/bin/${pname} \
        --chdir $out
  '';
}
