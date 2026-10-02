#!/bin/sh
# Выполняет проверки этапа 2 без открытия GUI; успешные GUI-сценарии проверяются вручную.
set -eu
cd "$(dirname "$0")/.."
make check
make test
make build
./build/cmd-emulator --help
sh tests/stage2-errors.sh
