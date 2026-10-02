#!/bin/sh
# Запускает приложение с JSON, включая заданный в нём стартовый скрипт.
set -eu
cd "$(dirname "$0")/.."
exec ./build/cmd-emulator --config configs/default.json
