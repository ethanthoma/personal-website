{ pkgs, odin }:

let
  ols = pkgs.ols.overrideAttrs (
    finalAttrs: previousAttrs: {
      buildInputs = [ odin ];
      env.ODIN_ROOT = "${odin}/share";
    }
  );
in
pkgs.mkShell {
  packages = [
    odin
    ols
  ];

  env.ODIN_ROOT = "${odin}/share";
}
