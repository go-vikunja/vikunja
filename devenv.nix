{ pkgs, lib, config, inputs, ... }:

let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
in {
  scripts.patch-sass-embedded.exec = ''
  find node_modules/.pnpm/sass-embedded-linux-*/node_modules/sass-embedded-linux-*/dart-sass/src -name dart -print0 | xargs -I {} -0 patchelf --set-interpreter "$(<$NIX_CC/nix-support/dynamic-linker)" {}
  '';

  packages = with pkgs-unstable; [
    # General tools
    git-cliff 
    actionlint
    crowdin-cli
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
      package = pkgs-unstable.go;
    };
  };
  
  services.mailpit = {
    enable = true;
    package = pkgs-unstable.mailpit;
  };
}
