#!/bin/sh
# Проверяет этапы 1–3 и отказ от запуска GUI при неверной VFS.
set -eu
cd "$(dirname "$0")/.."
sh tests/stage2.sh
status=0
output=$(./build/cmd-emulator --vfs tests/data/invalid-vfs.xml 2>&1) || status=$?
if [ "$status" -ne 1 ]; then
    printf 'Неверная VFS: ожидался код 1, получен %s\n' "$status" >&2
    exit 1
fi
case "$output" in
    *"Ошибка запуска:"*"vfs.LoadFile:"*"неверный base64"*) ;;
    *) printf 'Нет ожидаемой ошибки VFS: %s\n' "$output" >&2; exit 1 ;;
esac
printf 'OK: неверная VFS отклонена до создания окна\n'
