
#!/usr/bin/env bash

cd /source

export GOROOT=/opt/go
export GOPATH=/opt/.go
export PATH=${GOPATH}/bin:${GOROOT}/bin:${PATH}

# DARWIN BUILDS
export OSXCROSS="/opt/osxcross"
export SDK_VERSION=11.3
export DARWIN="${OSXCROSS}/target"
export DARWIN_SDK="${DARWIN}/SDK/MacOSX${SDK_VERSION}.sdk"

export PATH="${DARWIN}/bin:${DARWIN_SDK}/bin:${PATH}"
export LDFLAGS="-L${DARWIN_SDK}/lib -mmacosx-version-min=10.10"

export LD_FLAGS="-w -s -lSDL2main -lSDL2 -lSDL2_image -lSDL2_mixer -lSDL2_audio -lSDL2_ttf -lSDL -lsdl"

# echo "Compiling for linux/amd64..."
# GOOS=linux GOARCH=amd64 CGO_ENABLED=1  go build -tags static -ldflags "$LD_FLAGS" -o /build/gas-linux-amd64 main.go

echo "Compiling for linux/arm64..."
export PKG_CONFIG_PATH=/usr/aarch64-linux-gnu/lib/pkgconfig
CC=aarch64-linux-gnu-gcc-9 CXX=aarch64-linux-gnu-g++-9 GOOS=linux GOARCH=arm64 \
  CGO_ENABLED=1 go build -tags static -ldflags "$LD_FLAGS" -o /build/gas-linux-arm64 main.go

exit 0
echo "Compiling for darwin/arm64..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags static -ldflags "$LD_FLAGS" -o /build/gas-darwin.20.4-arm64 /source/main.go

echo "Compiling for linux/amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build --tags static --ldflags="${LD_FLAGS}" -o "/build/gas-linux-amd64" /source/main.go


# CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags static -ldflags "-s -w" -o /build/gas-darwin-arm64 /source/main.go
# CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -tags static -ldflags "-s -w" -o /build/gas-linux-arm64 /source/main.go
# CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -tags static -ldflags "-s -w" -o /build/gas-darwin-amd64 /source/main.go
# CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags static -ldflags "-s -w" -o /build/gas-linux-amd64 /source/main.go
# CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags static -ldflags "-s -w" -o /build/gas-windows-amd64 /source/main.go