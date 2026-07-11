{
  description = "label-maker – CSV → print-ready PDF label sheets";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        # ── Development shell ──────────────────────────────────────────────
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            golangci-lint
            gotools
          ];
        };

        # ── Buildable package ──────────────────────────────────────────────
        # To update after changing dependencies:
        #   1. Set vendorHash = pkgs.lib.fakeHash;
        #   2. Run `nix build` — it fails printing the correct sha256.
        #   3. Replace with that sha256.
        packages.default = pkgs.buildGoModule {
          pname = "label-maker";
          version = "0.1.0";
          src = ./.;

          vendorHash = "sha256-beRCLMXSJdb2N3UySsa/tTiuf+7MtBYmgVvCt3bT5bk=";

          meta = with pkgs.lib; {
            description = "Turn a CSV of addresses into a print-ready PDF of labels";
            license = licenses.mit;
            maintainers = [ ];
          };
        };

        # ── Runnable app ───────────────────────────────────────────────────
        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };
      });
}
