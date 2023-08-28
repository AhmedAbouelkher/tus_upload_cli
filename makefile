# https://stackoverflow.com/questions/2057689/how-does-make-app-know-default-target-to-build-if-no-target-is-specified
.DEFAULT_GOAL := build

.PHONY: default
default: build ;

build_windows:
	@echo "Building windows exe..."
	GOOS=windows go build -ldflags "-s -w" -o ./platforms/tus_upload.exe main.go
	@echo "Done."

build_linux:
	@echo "Building linux exe..."
	GOOS=linux go build -ldflags "-s -w" -o ./platforms/tus_upload main.go
	@echo "Done."

build_mac:
	@echo "Building mac exe..."
	GOOS=darwin go build -ldflags "-s -w" -o ./platforms/tus_upload_macos main.go
	@echo "Done."

build: build_windows build_linux build_mac