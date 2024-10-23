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
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
      templ,
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        gopkgs = gomod2nix.legacyPackages.${system};
        templpkgs = templ.packages.${system}.templ;

        webserverPort = "8080";

        callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
      in
      rec {
        packages.default = callPackage ./services/webserver {
          port = webserverPort;
          inherit (pkgs) makeWrapper tailwindcss;
          inherit (gopkgs) buildGoApplication;
          inherit templpkgs;
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
              pkgs.tailwindcss
              #    uploader
              templpkgs
            ];

            env.WEBSERVER_PORT = webserverPort;
          };

      }
    ));
}
