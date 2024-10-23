{ buildGoApplication }:
buildGoApplication rec {
  pname = "blob";
  version = "0.1.0";
  src = ../../.;
  pwd = ./.;
  modules = ../../gomod2nix.toml;
  subPackages = [ "services/${pname}" ];
}
