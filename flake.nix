{
  description = "Auxie dev environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            air        
            sqlite     
            gcc        

            bun
            
            gnumake    
          ];

          shellHook = ''
            echo "░░░▒▒▒▓▓██ WELCOME TO AUXIE DEV SHELL ██▓▓▒▒▒░░░"
            echo "┌ Go version: $(go version)"
            echo "└ Bun version: $(bun --version)"
          '';
        };
      });
}
