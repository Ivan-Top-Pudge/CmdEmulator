#!/bin/sh
# Загружает VFS с несколькими файлами и выполняет демонстрационный скрипт.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --vfs testdata/vfs/multiple.xml --script scripts/stage3-startup.txt
