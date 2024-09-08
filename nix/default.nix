{ 
pname, 
version,
pkgs ? (
    let
        inherit (builtins) fetchTree fromJSON readFile;
        inherit ((fromJSON (readFile ../flake.lock)).nodes) nixpkgs gomod2nix;
    in
        import (fetchTree nixpkgs.locked) {
            overlays = [
                (import "${fetchTree gomod2nix.locked}/overlay.nix")
            ];
        }
)
, buildGoApplication ? pkgs.buildGoApplication
}:

buildGoApplication {
    inherit pname version;
    src = ../.;
    pwd = ../.;
    modules = ../gomod2nix.toml;
    submodule = [ pname ];
    postInstall = ''
        cp -rf $src/static $out/static
        mkdir -p $out/cmd
        cp -rf $src/cmd/pages $out/cmd/pages
        cp -rf $src/cmd/components $out/cmd/components
    '';
}
