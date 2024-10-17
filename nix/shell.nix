{
  pkgs,
  mkGoEnv,
  gomod2nix,
  uploader,
  odin,
}:

let
  ols = pkgs.ols.overrideAttrs (
    finalAttrs: previousAttrs: {
      buildInputs = [ odin ];
      env.ODIN_ROOT = "${odin}/share";
    }
  );

  goEnv = mkGoEnv { pwd = ../.; };
in
pkgs.mkShell {
  packages = [
    goEnv
    gomod2nix
    pkgs.air
    pkgs.turso-cli
    uploader

    odin
    ols
  ];

  env.ODIN_ROOT = "${odin}/share";
}
