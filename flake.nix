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
        # vendor/ is committed, so vendorHash = null tells buildGoModule to use it.
        # After adding new dependencies:
        #   nix develop --command go mod tidy
        #   nix develop --command go mod vendor
        #   git add vendor
        packages.default = pkgs.buildGoModule {
          pname = "label-maker";
          version = "0.1.0";
          src = ./.;

          # vendor/ is checked in, so Nix uses it directly.
          vendorHash = null;

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
