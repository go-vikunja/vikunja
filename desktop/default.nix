{ pkgs ? import <nixpkgs> {}
}:
pkgs.mkShell {
  name="electron-dev";
  buildInputs = [
    pkgs.electron
  ];
}

