{ 
pname, 
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
    inherit pname;
    version = "0.1";
    src = ../.;
    pwd = ../.;
    modules = ../gomod2nix.toml;
    submodule = [ pname ];
}
