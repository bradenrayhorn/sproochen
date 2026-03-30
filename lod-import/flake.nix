{
  description = "import app environment";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";

  outputs =
    { self, nixpkgs }:
    let
      pkgs = nixpkgs.legacyPackages.aarch64-linux;
    in
    {
      devShells.aarch64-linux.default = pkgs.mkShell {
        name = "go";
        packages = [
          pkgs.go
        ];
      };
    };
}
