{
  description = "🍹 Juice - Brazilian NFC-e Invoice Extraction Library";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      systems = [
        "aarch64-linux"
        "x86_64-linux"
      ];

      forEachSystem = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forEachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "juice";
            version = "alpha";
            src = ./.;
            vendorHash = "sha256-PmOUK4yXy8J18YNsChxZ5xzUEsgZL6LDMumA3tGQzNE=";

            meta = with pkgs.lib; {
              description = "Juice - Brazilian NFC-e Invoice Extraction Library";
              homepage = "https://github.com/glwbr/juice";
              license = licenses.mit;
              mainProgram = "juice";
            };
          };
        }
      );

      devShells = forEachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              golangci-lint
              treefmt
              nixfmt-rfc-style
            ];
            shellHook = ''echo "A juicy environment is ready!"'';
          };
        }
      );

      apps = forEachSystem (system: {
        default = {
          type = "app";
          program = "${nixpkgs.legacyPackages.${system}.lib.getExe self.packages.${system}.default}";
        };
      });

      defaultPackage = forEachSystem (system: self.packages.${system}.default);
      defaultApp = forEachSystem (system: self.apps.${system}.default);
    };
}
