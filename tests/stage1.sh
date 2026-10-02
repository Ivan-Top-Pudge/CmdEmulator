#!/bin/sh
# Проверки и сборка этапа 1 без запуска графического окна.
set -eu

cd "$(dirname "$0")/.."
make check
make test
make build
