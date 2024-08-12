{ pkgs, lib, config, inputs, ... }:

let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
in {
  packages = with pkgs; [
    # General tools
	git-cliff 
    # Frontend tools
    cypress 
    # API tools
    golangci-lint mage
    # Desktop
    electron
  ];
  
  languages = {
    javascript = {
      enable = true;
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
