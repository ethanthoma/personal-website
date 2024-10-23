{
  fetchurl,
  makeWrapper,
  buildGoApplication,
  tailwindcss,
  templpkgs,
  port,
}:
let
  htmxVersion = "2.0.2";
  htmx = fetchurl {
    url = "https://github.com/bigskysoftware/htmx/releases/download/v${htmxVersion}/htmx.min.js";
    hash = "sha256-4XRtl1nsDUPFwoRFIzOjELtf1yheusSy3Jv0TXK1qIc=";
  };
in
buildGoApplication rec {
  pname = "webserver";
  version = "0.1";
  src = ../../.;
  pwd = ./.;
  modules = ../../gomod2nix.toml;
  subPackages = [ "services/${pname}" ];
  env.port = port;
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

    tailwindcss -c ./services/${pname}/tailwind.config.js -i $public/main.css -o $static/main.css --minify

    rsync -a $public $out --exclude js --exclude='*.css'

    mkdir -p $static/js
    cp ${htmx} $static/js/htmx.min.js

    mv $out/bin/${pname} $out/bin/.${pname}-unwrapped
    makeWrapper $out/bin/.${pname}-unwrapped $out/bin/${pname} \
        --set WEBSERVER_PORT "${port}" \
        --chdir $out
  '';
}
