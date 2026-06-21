{
  description = "Generated Go bindings for the Nix C API";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs systems;

      perSystem =
        system:
        let
          pkgs = import nixpkgs { inherit system; };
          lib = pkgs.lib;

          nixCLibs = with pkgs.nixVersions.latest.libs; [
            nix-util-c
            nix-store-c
            nix-expr-c
            nix-fetchers-c
            nix-flake-c
            nix-main-c
          ];

          nixCppLibs = with pkgs.nixVersions.latest.libs; [
            nix-flake
          ];

          nixLibs = nixCLibs ++ nixCppLibs;

          pkgConfigPath = lib.makeSearchPath "lib/pkgconfig" (map lib.getDev nixLibs);

          generateGoBindingsTool =
            pkgs.writeShellApplication {
              name = "generate-go-bindings";

              runtimeInputs = [
                pkgs.c-for-go
                pkgs.pkg-config
                pkgs.coreutils
              ];

              runtimeEnv = {
                PKG_CONFIG_PATH = pkgConfigPath;
              };

              text = ''
                set -euo pipefail

                repo="''${1:-$PWD}"
                cd "$repo"

                if [ ! -f nix-go-bindings.yml ]; then
                  echo "generate-go-bindings must run from the repository root" >&2
                  exit 1
                fi

                tmp="$(mktemp -d)"
                trap 'rm -rf "$tmp"' EXIT HUP INT TERM

                c-for-go \
                  -nostamp \
                  -out "$tmp/out" \
                  nix-go-bindings.yml

                cp -r "$tmp/out/nix/." .
              '';
            };
        in
        {
          inherit
            pkgs
            nixCLibs
            nixCppLibs
            nixLibs
            pkgConfigPath
            generateGoBindingsTool
            ;
        };
    in
    {
      packages = forAllSystems (
        system:
        let
          env = perSystem system;
        in
        {
          default = env.generateGoBindingsTool;
          generate-go-bindings-tool = env.generateGoBindingsTool;
        }
      );

      apps = forAllSystems (
        system:
        {
          generate-go-bindings = {
            type = "app";
            program = "${self.packages.${system}.generate-go-bindings-tool}/bin/generate-go-bindings";
          };

          default = self.apps.${system}.generate-go-bindings;
        }
      );

      devShells = forAllSystems (
        system:
        let
          env = perSystem system;
        in
        {
          default = env.pkgs.mkShell {
            packages = [
              env.pkgs.go
              env.pkgs.c-for-go
              env.pkgs.pkg-config
              env.pkgs.golangci-lint
            ] ++ env.nixLibs;

            shellHook = ''
              export CGO_ENABLED=1
              export CGO_CFLAGS_ALLOW=--isystem.*
              export CGO_CXXFLAGS_ALLOW=--isystem.*
              export PKG_CONFIG_PATH="${env.pkgConfigPath}''${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
            '';
          };
        }
      );
    };
}
