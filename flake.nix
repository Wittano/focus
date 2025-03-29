{
  description = "A simple Go package";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ.url = "github:a-h/templ";
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { inherit system; overlays = [templ.overlays.default]; };
        in
        {
          package.default = pkgs.callPackage ./default.nix { };
          devShells.default = pkgs.callPackage ./shell.nix { };
        });
}
