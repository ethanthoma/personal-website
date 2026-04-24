{
  fetchurl,
  makeWrapper,
  buildGoApplication,
  tailwindcss,
  templpkgs,
}:
let
  datastarVersion = "1.0.1";
  datastar = fetchurl {
    url = "https://raw.githubusercontent.com/starfederation/datastar/v${datastarVersion}/bundles/datastar.js";
    hash = "sha256-VHaM80mFvgIpxyKfHflGn70y4qDAm0o/HoGtjE1oQNo=";
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
    cp ${datastar} $static/js/datastar.js

    cp -r ./services/${pname}/pages $out/

    mv $out/bin/${pname} $out/bin/.${pname}-unwrapped
    makeWrapper $out/bin/.${pname}-unwrapped $out/bin/${pname} \
        --chdir $out
  '';
}
