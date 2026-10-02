#!/bin/sh
# Переопределяет приглашение JSON через CLI и отключает стартовый скрипт.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --config configs/default.json --prompt 'override> ' --script ''
