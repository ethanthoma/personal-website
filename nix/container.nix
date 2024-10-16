{ pkgs
, derivation
}:
pkgs.dockerTools.buildImage {
  name = derivation.pname;
  tag = derivation.version;
  created = "now";
  copyToRoot = pkgs.buildEnv {
    name = "image-root";
    paths = [ derivation ];
    pathsToLink = [ "/bin" "/cmd/${derivation.pname}" "/cmd/${derivation.pname}" "/static" ];
  };
  config = {
    Cmd = [
      "${derivation}/bin/${derivation.pname}"
    ];
    Env = [
      "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
    ];
    ExposedPorts = {
      "8080/tcp" = { };
    };
  };
}

