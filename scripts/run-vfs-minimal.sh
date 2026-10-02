#!/bin/sh
# Загружает минимальную VFS и выполняет демонстрационный скрипт; ожидает ручной ввод.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --vfs testdata/vfs/minimal.xml --script scripts/stage3-startup.txt
