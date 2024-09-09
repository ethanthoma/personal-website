{ pname
, version
, pkgs
, buildGoApplication
}:

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
        mkdir -p $out/static/styles

        rsync -a $src/static $out --exclude styles
        lightningcss --bundle $src/static/styles/main.css -t "> .5% or last 2 versions" -o $out/static/styles/main.css

        mkdir -p $out/cmd/${pname}
        cp -rf $src/cmd/${pname}/pages $out/cmd/${pname}/pages
        cp -rf $src/cmd/${pname}/components $out/cmd/${pname}/components
    '';
}
