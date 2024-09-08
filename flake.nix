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
                    pname = "cmd";
                    version = "0.1";

                    pkgs = nixpkgs.legacyPackages.${system};

                    # The current default sdk for macOS fails to compile go projects, so we use a newer one for now.
                    # This has no effect on other platforms.
                    callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
                in
                    rec {
                    packages.default = callPackage ./nix {
                        inherit pname version;
                        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
                    };

                    packages.container = pkgs.dockerTools.buildImage {
                        name = pname;   
                        tag = version;
                        created = "now";
                        copyToRoot = pkgs.buildEnv {
                            name = "image-root";
                            paths = [ packages.default  ];
                            pathsToLink = [ "/bin" "/cmd/components" "/cmd/pages" "/static" ];
                        };
                        config = {
                            Cmd = [ 
                                "${packages.default}/bin/${pname}" 
                            ];
                            Env = [
                                "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
                            ];
                            ExposedPorts = {
                                "8080/tcp" = {};
                            };
                        };
                    };

                    devShells.default = callPackage ./nix/shell.nix {
                        env.GOFLAGS = "-mod=vendor";
                        inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
                    };
                })
        );
}
