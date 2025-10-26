rm -f cmd/template/.DS_Store

CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/macos_arm/zstdfs -ldflags "-w -s" main.go
zip dist/macos_arm/zstdfs_macos_arm.zip dist/macos_arm/zstdfs

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/macos_intel/zstdfs -ldflags "-w -s" main.go
zip dist/macos_intel/zstdfs_macos_intel.zip dist/macos_intel/zstdfs


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/linux_amd64/zstdfs -ldflags "-w -s" main.go
zip dist/linux_amd64/zstdfs_linux_amd64.zip dist/linux_amd64/zstdfs


CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/windows_amd64/zstdfs.exe -ldflags "-w -s" main.go
zip dist/windows_amd64/zstdfs_windows_amd64.zip dist/windows_amd64/zstdfs.exe
