# Claude Swap Widget — build & package
#
# Targets:
#   make backend         build csw binary  -> backend/bin/csw
#   make widget          build SwiftPM executable
#   make app             bundle .app with csw embedded -> release/ClaudeSwapWidget.app
#   make install         copy .app to /Applications, csw to /usr/local/bin
#   make run             run widget directly (dev)
#   make test            go vet + swift test
#   make clean           wipe dist + .build + backend/bin

SHELL := /bin/bash
APP_NAME := ClaudeSwapWidget
BUNDLE_ID := dev.soi.claude-swap-widget
BACKEND_BIN := backend/bin/csw
WIDGET_BUILD := widget/.build/release/$(APP_NAME)
APP_BUNDLE := release/$(APP_NAME).app

.PHONY: all backend widget app install run test clean

all: app

backend:
	@mkdir -p backend/bin
	cd backend && go build -trimpath -ldflags="-s -w" -o bin/csw ./cmd/csw

widget:
	cd widget && swift build -c release

app: backend widget
	@rm -rf $(APP_BUNDLE)
	@mkdir -p $(APP_BUNDLE)/Contents/MacOS
	@mkdir -p $(APP_BUNDLE)/Contents/Resources
	cp $(WIDGET_BUILD) $(APP_BUNDLE)/Contents/MacOS/$(APP_NAME)
	cp $(BACKEND_BIN) $(APP_BUNDLE)/Contents/Resources/csw
	cp packaging/icon.png $(APP_BUNDLE)/Contents/Resources/icon.png
	cp packaging/AppIcon.icns $(APP_BUNDLE)/Contents/Resources/AppIcon.icns
	cp packaging/Info.plist $(APP_BUNDLE)/Contents/Info.plist
	codesign --force --deep --sign "ClaudeSwapWidgetLocalDev" $(APP_BUNDLE) 2>/dev/null || codesign --force --deep --sign - $(APP_BUNDLE)
	@echo "Built $(APP_BUNDLE)"

install: app
	@rm -rf /Applications/$(APP_NAME).app
	cp -R $(APP_BUNDLE) /Applications/
	codesign --force --deep --sign "ClaudeSwapWidgetLocalDev" /Applications/$(APP_NAME).app 2>/dev/null || codesign --force --deep --sign - /Applications/$(APP_NAME).app
	@mkdir -p /usr/local/bin 2>/dev/null || true
	cp $(BACKEND_BIN) /usr/local/bin/csw
	@echo "Installed to /Applications/$(APP_NAME).app and /usr/local/bin/csw"

run: backend
	cd widget && CSW_BIN=$(PWD)/$(BACKEND_BIN) swift run

test:
	cd backend && go vet ./... && go test ./...
	cd widget && swift test

clean:
	rm -rf release backend/bin widget/.build
