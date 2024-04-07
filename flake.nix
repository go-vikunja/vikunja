{
  description = "Vikunja dev environment";

  outputs = { self, nixpkgs }:
    let pkgs = nixpkgs.legacyPackages.x86_64-linux;
    in {
      defaultPackage.x86_64-linux =
        pkgs.mkShell { buildInputs = with pkgs; [
          # General tools
          git-cliff 
          # Frontend tools
          nodePackages.pnpm cypress 
          # API tools
          go golangci-lint mage
        ]; 
      };
    };
}
