#!/bin/sh
# Открывает демонстрацию изменения VFS в памяти.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --vfs testdata/vfs/commands.xml --script scripts/stage5-startup.txt
