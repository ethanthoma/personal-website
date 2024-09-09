{ pkgs 
, mkGoEnv
, gomod2nix
, uploader
, env ? {}
}:

let
    goEnv = mkGoEnv { pwd = ../.; };
in
    pkgs.mkShell {
        inherit env;
        packages = [
            goEnv
            gomod2nix
            pkgs.air
            pkgs.turso-cli
            uploader
        ];
    }
