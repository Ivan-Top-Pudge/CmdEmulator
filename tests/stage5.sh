#!/bin/sh
# Проверяет этапы 1–5, включая сохранность XML после mkdir и chmod.
set -eu
cd "$(dirname "$0")/.."
sh tests/stage4.sh
