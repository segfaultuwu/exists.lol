{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  name = "exists-lol-dev";

  packages = with pkgs; [
    # Go
    go
    gopls
    gotools
    go-tools
    delve

    # Build / dev tools
    git
    gnumake
    curl
    jq
    ripgrep
    fd

    # SQLite
    sqlite
    sqlc

    # Env files
    dotenv-linter
  ];

  shellHook = ''
    echo "exists.lol dev shell"
    echo "Go: $(go version)"
    echo ""

    export CGO_ENABLED=1

    mkdir -p bin

    alias run="go run ./cmd/existsbot"
    alias test="go test ./..."
    alias tidy="go mod tidy"
    alias fmt="go fmt ./..."
    alias vet="go vet ./..."
    alias check="go fmt ./... && go vet ./... && go test ./..."

    build() {
      mkdir -p bin

      go build -o bin/existsbot ./cmd/existsbot &&
      go build -o bin/existslol ./cmd/existslol
    }
  '';
}
