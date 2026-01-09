{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "ctfbot";
          version = "0.1.0";
          src = ./.;

          # Run 'nix build' and it will fail with the correct hash if this is wrong.
          # The user can then update it here.
          vendorHash = "sha256-0000000000000000000000000000000000000000000=";

          subPackages = [ "cmd/ctfbotd" ];
          
          nativeBuildInputs = [ pkgs.installShellFiles ];
          
          # Ensure we can find the sqlite dynamic library if needed, 
          # though modernc.org/sqlite is a pure Go implementation.
          buildInputs = [ pkgs.sqlite ];
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            sqlite-interactive
          ];

          shellHook = ''
            echo "Welcome to the CTFBot development environment!"
            echo "Go version: $(go version)"
            echo "You can start the bot with: go run ./cmd/ctfbotd"
          '';
        };
      }
    );
}
