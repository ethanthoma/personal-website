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
}
