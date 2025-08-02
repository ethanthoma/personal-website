{
  description = "Go Webserver Flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devshell.url = "github:numtide/devshell";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs = {
        nixpkgs.follows = "nixpkgs";
      };
    };
    templ.url = "github:a-h/templ";
  };

  outputs =
    inputs@{ self, ... }:

    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.devshell.flakeModule ];

      systems = [ "x86_64-linux" ];

      perSystem =
        { system, ... }:

        let
          pkgs = import inputs.nixpkgs {
            system = "x86_64-linux";
            config.allowUnfree = true;
            overlays = [
              inputs.gomod2nix.overlays.default
            ];
          };

          templpkgs = inputs.templ.packages.${system}.templ;
        in
        {
          _module.args.pkgs = pkgs;

          devshells.default = {
            packages = [
              pkgs.air
              pkgs.turso-cli
              pkgs.gopls
              pkgs.tailwindcss_4
              pkgs.watchman
              pkgs.tailwindcss-language-server
              templpkgs
              pkgs.mdformat
              pkgs.rustywind
              pkgs.stylelint
              pkgs.biome
              pkgs.mago
              pkgs.superhtml
            ];

            commands = [
              {
                name = "claude";
                package = pkgs.claude-code;
              }
              {
                name = "make";
                package = pkgs.gnumake;
              }
              {
                package = pkgs.gomod2nix;
              }
              {
                package = pkgs.go;
              }
            ];

            env = [
              {
                name = "WEBSERVER_PORT";
                value = "8080";
              }
            ];
          };

          packages.default = pkgs.callPackage ./services/webserver {
            inherit (pkgs) makeWrapper buildGoApplication;
            inherit templpkgs;
            tailwindcss = pkgs.tailwindcss_4;
          };
        };
    };
}
