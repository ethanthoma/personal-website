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

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;

        odin = callPackage ./nix/odin.nix {
          MacOSX-SDK = pkgs.darwin.apple_sdk;
          inherit (pkgs.darwin) Security;
        };
      in
      rec {
        packages.default = callPackage ./nix/webserver.nix {
          pname = "webserver";
          version = "0.1";
          inherit pkgs odin;
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
          uploader = packages.uploader;

          inherit odin;
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        };
      }
    ));
}
