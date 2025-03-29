{ mkShell
, go_1_24
, gotools
, gopls
, act
, nixd
, nixpkgs-fmt
, templ
}: mkShell {
  hardeningDisable = [ "all" ];

  GOROOT = "${go_1_24}/share/go";
  DEBUG_ARGS = "";

  nativeBuildInputs = [
    # GO
    go_1_24
    gotools
    gopls
    templ

    # Nix
    nixpkgs-fmt
    nixd

    # Github actions
    act
  ];
}
