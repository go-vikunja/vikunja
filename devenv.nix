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
    nfpm
		pkgs-unstable.watchexec
    # API tools
    golangci-lint mage
    # Desktop
    electron
    # Font processing tools
    wget
    python3
    python3Packages.pip
    python3Packages.fonttools
    python3Packages.brotli
  ] ++ lib.optionals (!pkgs.stdenv.isDarwin) [
    # Frontend tools (exclude on Darwin)
    pkgs-unstable.cypress
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
			enableHardeningWorkaround = true;
    };
  };
  
  services.mailpit = {
    enable = true;
    package = pkgs-unstable.mailpit;
  };

	# Global environment variables
	env = {
    VIKUNJA_SERVICE_FRONTENDURL = "http://localhost:4173";
    VIKUNJA_DATABASE_TYPE = "sqlite";
    VIKUNJA_DATABASE_PATH = "/tmp/vikunja.db";
		VIKUNJA_SERVICE_INTERFACE = "127.0.0.1:3456";
  };

  # Starts the API and frontend
  processes = {
    api = {
      #exec = "mage build && ./vikunja";
      exec = "watchexec -r -e go -- 'go run .'";
    };
    frontend = {
      exec = "pnpm --dir frontend run serve";
    };
  };

	devcontainer = {
		enable = true;
		settings = {
			forwardPorts = [ 4173 3456 ];
			portsAttributes = {
				"4173" = {
					label = "Vikunja Frontend dev server";
				};
				"3456" = {
					label = "Vikunja API";
				};
			};
			customizations.vscode.extensions = [
        "Syler.sass-indented"
        "codezombiech.gitignore"
        "dbaeumer.vscode-eslint"
        "editorconfig.editorconfig"
        "golang.Go"
        "lokalise.i18n-ally"
        "mikestead.dotenv"
        "mkhl.direnv"
        "vitest.explorer"
        "vue.volar"
			];
		};
	};
}
