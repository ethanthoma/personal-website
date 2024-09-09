{ pname
, version
, buildGoApplication
}:

buildGoApplication {
    inherit pname version;
    src = ../.;
    pwd = ../.;
    modules = ../gomod2nix.toml;
    subPackages = [ "cmd/${pname}" ];
    postInstall = ''
        cp -rf $src/static $out/static
        mkdir -p $out/cmd/${pname}
        cp -rf $src/cmd/${pname}/pages $out/cmd/${pname}/pages
        cp -rf $src/cmd/${pname}/components $out/cmd/${pname}/components
    '';
}
