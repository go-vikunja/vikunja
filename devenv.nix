{ pkgs, lib, config, inputs, ... }:

let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
  # nixpkgs still ships golangci-lint 2.12.2 built with go 1.26, which refuses
  # a go.mod targeting 1.27; drop when nixpkgs bumps to >= 2.13.0 on go 1.27
  golangci-lint-go127 =
    (pkgs-unstable.golangci-lint.override {
      buildGo126Module = pkgs-unstable.buildGo127Module;
    }).overrideAttrs (finalAttrs: prev: {
      version = "2.13.0";
      src = pkgs-unstable.fetchFromGitHub {
        owner = "golangci";
        repo = "golangci-lint";
        tag = "v${finalAttrs.version}";
        hash = "sha256-WWKvf1uQr8QK0Ja+EjN0YbLMX27N23HvyfFFKfjQ1gg=";
      };
      vendorHash = "sha256-thmkiCuE4FnVTIExfwN7xm6xioxz4C+tagvIsre/s5A=";
    });
in {
  # staticcheck 2026.1 tests panic when rebuilt with go 1.27; drop when nixpkgs ships a fixed release
  overlays = [
    (final: prev: { go-tools = prev.go-tools.overrideAttrs (_: { doCheck = false; }); })
  ];

  scripts.patch-sass-embedded.exec = ''
  find node_modules/.pnpm/sass-embedded-linux-*/node_modules/sass-embedded-linux-*/dart-sass/src -name dart -print0 | xargs -I {} -0 patchelf --set-interpreter "$(<$NIX_CC/nix-support/dynamic-linker)" {}
  '';

	packages = with pkgs-unstable; [
    # General tools
    git-cliff 
    actionlint
    crowdin-cli
    nfpm
    # API tools
    golangci-lint-go127 mage
    # Desktop
    electron
    # Font processing tools
    wget
    python3
    python3Packages.pip
    python3Packages.fonttools
    python3Packages.brotli
    nodejs
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
      package = pkgs-unstable.go_1_27;
      enableHardeningWorkaround = true;
    };
  };
  
  services.mailpit = {
    enable = true;
    package = pkgs-unstable.mailpit;
  };

  env = {
    PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD = "1";
    PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS = "1";
#    PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH = "${pkgs-unstable.chromium}/bin/chromium";
    VIKUNJA_SERVICE_TESTINGTOKEN = "test";
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
