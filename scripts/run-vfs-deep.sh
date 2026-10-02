#!/bin/sh
# Загружает VFS с вложенными каталогами и выполняет демонстрационный скрипт.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --vfs testdata/vfs/deep.xml --script scripts/stage3-startup.txt
