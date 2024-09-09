{
    description = "A basic gomod2nix flake";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
        flake-utils.url = "github:numtide/flake-utils";
        gomod2nix = {
            url = "github:nix-community/gomod2nix";
            inputs = {
                nixpkgs.follows = "nixpkgs";
                flake-utils.follows = "flake-utils";
            };
        };
    };

    outputs = { self, nixpkgs, flake-utils, gomod2nix }:
        (flake-utils.lib.eachDefaultSystem
            (system:
                let
                    pkgs = nixpkgs.legacyPackages.${system};

                    # The current default sdk for macOS fails to compile go projects, so we use a newer one for now.
                    # This has no effect on other platforms.
                    callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
                in
                    rec {
                    packages.default = callPackage ./nix/webserver.nix {
                        pname = "webserver";
                        version = "0.1";
                        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
                    };

                    packages.uploader = callPackage ./nix/uploader.nix {
                        pname = "uploader";
                        version = "0.1";
                        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
                    };

                    packages.container = callPackage ./nix/container.nix {
                        derivation = packages.default;
                        inherit pkgs;
                    };

                    devShells.default = callPackage ./nix/shell.nix {
                        env.GOFLAGS = "-mod=vendor";
                        uploader = packages.uploader;
                        inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
                    };
                })
        );
}
