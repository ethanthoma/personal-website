{
  description = "Go Webserver Flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devshell.url = "github:numtide/devshell";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    templ = {
      url = "github:a-h/templ?ref=v0.3.1001";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs:

    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.devshell.flakeModule ];

      systems = [ "x86_64-linux" ];

      perSystem =
        { system, pkgs, ... }:

        let
          templpkgs = inputs.templ.packages.${system}.templ;
          gomod2nixpkgs = inputs.gomod2nix.packages.${system}.default;
          buildGoApplication = inputs.gomod2nix.legacyPackages.${system}.buildGoApplication;
        in
        {
          devshells.default = {
            packages = [
              pkgs.air
              pkgs.turso-cli
              pkgs.gopls
              pkgs.tailwindcss_4
              pkgs.tailwindcss-language-server
              templpkgs

              pkgs.mdformat
              pkgs.rustywind
              pkgs.stylelint
              pkgs.biome
              pkgs.mago
              pkgs.superhtml

              pkgs.google-lighthouse
              pkgs.chromium
              pkgs.jq
            ];

            commands = [
              { package = gomod2nixpkgs; }
              { package = pkgs.go; }
            ];

            env = [
              {
                name = "WEBSERVER_PORT";
                value = "8080";
              }
              {
                name = "CGO_ENABLED";
                value = "0";
              }
            ];
          };

          packages.default = pkgs.callPackage ./services/webserver {
            inherit templpkgs buildGoApplication;
            tailwindcss = pkgs.tailwindcss_4;
          };
        };
    };
}
