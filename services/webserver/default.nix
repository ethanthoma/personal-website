{
  fetchurl,
  makeWrapper,
  python3,
  rsync,
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
  pythonFonttools = python3.withPackages (p: [
    p.fonttools
    p.brotli
  ]);
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
    pythonFonttools
    rsync
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

    # Subset shipped fonts: pin Monaspace's unused width/slant axes (kept wght
    # only), then strip non-Latin glyphs and unused OpenType features. Cuts
    # Monaspace Neon ~184 KB → ~28 KB and Krypton ~148 KB → ~22 KB. Public Sans
    # has only a wght axis so just gets glyph/feature subsetting.
    UNICODES='U+0020-007E,U+00A0-00FF,U+2010-206F,U+2190-21FF'
    FEATURES='kern,liga,calt,ccmp,locl'
    subset_var() {
      local f=$1
      fonttools varLib.instancer "$f" wdth=100 slnt=0 -o "$f.tmp"
      pyftsubset "$f.tmp" --output-file="$f" \
        --unicodes="$UNICODES" --layout-features="$FEATURES" \
        --flavor=woff2 --no-hinting --no-glyph-names
      rm "$f.tmp"
    }
    subset_static() {
      local f=$1
      pyftsubset "$f" --output-file="$f" \
        --unicodes="$UNICODES" --layout-features="$FEATURES" \
        --flavor=woff2 --no-hinting --no-glyph-names
    }
    subset_var "$static/fonts/Monaspace/MonaspaceNeonVarVF[wght,wdth,slnt].woff2"
    subset_var "$static/fonts/Monaspace/MonaspaceKryptonVarVF[wght,wdth,slnt].woff2"
    subset_static "$static/fonts/PublicSans/PublicSans[wght].woff2"
    subset_static "$static/fonts/PublicSans/PublicSans-Italic[wght].woff2"

    cp -r ./services/${pname}/pages $out/

    mv $out/bin/${pname} $out/bin/.${pname}-unwrapped
    makeWrapper $out/bin/.${pname}-unwrapped $out/bin/${pname} \
        --chdir $out
  '';
}
