#!/bin/sh
# Запускает приложение со стартовым скриптом, заданным через CLI.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --vfs testdata/stage2-vfs.xml --prompt 'script> ' --script scripts/startup.txt
