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
    templ.url = "github:a-h/templ";
  };

  outputs =
    inputs@{ self, ... }:

    (inputs.flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import inputs.nixpkgs {
          system = "x86_64-linux";
          config.allowUnfree = true;
        };

        gopkgs = inputs.gomod2nix.legacyPackages.${system};
        templpkgs = inputs.templ.packages.${system}.templ;

        webserverPort = "8080";

        callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
      in
      rec {
        packages.default = callPackage ./services/webserver {
          port = webserverPort;
          inherit (pkgs) makeWrapper;
          inherit (gopkgs) buildGoApplication;
          inherit templpkgs;
          tailwindcss = pkgs.tailwindcss_4;
        };

        packages.uploader = callPackage ./cmd/uploader { inherit (gopkgs) buildGoApplication; };

        packages.blob = callPackage ./services/blob { inherit (gopkgs) buildGoApplication; };

        packages.container = pkgs.dockerTools.buildImage {
          name = packages.default.pname;
          tag = packages.default.version;
          created = "now";
          copyToRoot = pkgs.buildEnv {
            name = "image-root";
            paths = [ packages.default ];
            pathsToLink = [
              "/bin"
              "/public"
            ];
          };
          config = {
            Cmd = [ "${packages.default}/bin/${packages.default.pname}" ];
            Env = [ "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt" ];
            ExposedPorts = {
              "${webserverPort}/tcp" = { };
            };
          };
        };

        devShells.default =
          let
            goEnv = gopkgs.mkGoEnv { pwd = ./.; };
          in
          pkgs.mkShell {
            packages = [
              goEnv
              gopkgs.gomod2nix
              pkgs.air
              pkgs.turso-cli
              pkgs.gopls
              pkgs.tailwindcss_4
              pkgs.watchman
              pkgs.tailwindcss-language-server
              packages.uploader
              templpkgs
              pkgs.mdformat
              pkgs.rustywind
              pkgs.stylelint
              pkgs.biome
              pkgs.mago
              pkgs.superhtml
              pkgs.claude-code
            ];

            env.WEBSERVER_PORT = webserverPort;
          };
      }
    ));
}
