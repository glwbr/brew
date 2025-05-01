{
  description = "🌬️ Brisa -  Brazilian Invoice Scraping Agent";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";

  outputs = { self, nixpkgs, ... }:
    let 
      systems = [ "aarch64-linux" "x86_64-linux" ];
      forEachSystem = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forEachSystem (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "brisa";
            version = "alpha";
            src = ./.;
            vendorHash = "sha256-Ifh8WbCPFFgyF615UwIbqriTNZJJk2HXD9BlkcMJgZM=";
            meta = with pkgs.lib; {
              description = "Library for extracting data from Brazilian electronic tax receipts (NFC-e)";
              homepage = "https://github.com/glwbr/brisa";
              license = licenses.mit;
              mainProgram = "brisa";
            };
          };
        }
      );

      devShells = forEachSystem (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go treefmt nixfmt-rfc-style nodePackages.prettier ];
            shellHook = ''
              export BRISA_DEV=1
              echo "🌬️ Brisa development environment loaded"
              echo "Type 'go run cmd/brisa/main.go or nix run' to start the application"
            '';
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
