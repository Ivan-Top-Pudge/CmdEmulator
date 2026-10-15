#!/bin/sh
# Открывает демонстрацию основных команд в GUI.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --vfs testdata/vfs/commands.xml --script scripts/stage4-startup.txt
