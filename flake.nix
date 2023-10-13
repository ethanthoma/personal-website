{
    description = "A flake for building and testing a Bazel project";

    inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    outputs = { self, nixpkgs }: {

        devShell.x86_64-linux = with nixpkgs.legacyPackages.x86_64-linux; mkShell {
            buildInputs = [
                bazel
                pulumi
            ];

            shellHook = ''
#                export TEST_SRCDIR=$PWD
#                export TEST_TMPDIR=/tmp
                export HOME=$PWD
            '';
        };
    };
}

