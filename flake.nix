{
    description = "Dev env for building, testing, and deploying my website";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
        flake-utils.url = "github:numtide/flake-utils";
    };

    outputs = { self, nixpkgs, flake-utils }:

    flake-utils.lib.eachDefaultSystem (system:
        let
            pkgs = import nixpkgs {
                inherit system;
            };
        in {
            devShell = pkgs.mkShell {
                buildInputs = [
                    pkgs.pulumi
                    pkgs.awscli2
                ];

                shellHook = '' '';
            };
        }
    );
}

