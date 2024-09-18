{ pkgs, lib, config, inputs, ... }:

let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
in {
  packages = with pkgs-unstable; [
    # General tools
    git-cliff 
    # API tools
    golangci-lint mage
    # Desktop
    electron
  ] ++ lib.optionals (!pkgs.stdenv.isDarwin) [
    # Frontend tools (exclude on Darwin)
    cypress
  ];
  
  languages = {
    javascript = {
      enable = true;
      package = pkgs-unstable.nodejs-slim;
      pnpm = {
        enable = true;
        package = pkgs-unstable.pnpm;
      };
    };
    
    go = {
      enable = true;
    };
  };
}
