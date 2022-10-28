{
  description = "Vikunja frontend dev environment";

  outputs = { self, nixpkgs }:
    let pkgs = nixpkgs.legacyPackages.x86_64-linux;
    in {
      defaultPackage.x86_64-linux =
        pkgs.mkShell { buildInputs = [ pkgs.nodePackages.pnpm pkgs.cypress pkgs.git-cliff ]; };
    };
}
