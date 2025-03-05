{ mkShell
, go_1_24
, gotools
, act
, nixd
, nixpkgs-fmt
}: mkShell {
  hardeningDisable = [ "all" ];

  GOROOT = "${go_1_24}/share/go";
  DEBUG_ARGS = "";

  nativeBuildInputs = [
    # GO
    go_1_24
    gotools

    # Nix
    nixpkgs-fmt
    nixd

    # Github actions
    act
  ];
}
