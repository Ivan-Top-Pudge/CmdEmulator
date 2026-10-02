#!/bin/sh
# Запускает приложение только с параметрами CLI; ожидает ручной ввод.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --vfs testdata/stage2-vfs.xml --prompt 'cli> '
