{
  description = "Go project";

  outputs =
    {
      self,
      nixpkgs,
      ...
    }:
    let
      inherit (nixpkgs) lib;
      eachSystem =
        f:
        lib.genAttrs nixpkgs.lib.systems.flakeExposed (system: f system nixpkgs.legacyPackages.${system});

    in
    {
      packages = eachSystem (system: pkgs: { });

      devShells = eachSystem (
        system: pkgs: {
          default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.gopls
              pkgs.gotools
              pkgs.cobra-cli
            ];
          };
        }
      );
    };
}
