{ buildGoApplication }:

buildGoApplication rec {
  pname = "uploader";
  version = "1.0";
  src = ../../.;
  pwd = ./.;
  modules = ../../gomod2nix.toml;
  subPackages = [ "cmd/${pname}" ];
}
