#!/bin/sh
# Проверяет код 1 и сообщение об ошибке для неверных параметров, JSON и файлов.
set -eu
cd "$(dirname "$0")/.."

# expect_failure запускает приложение и проверяет ожидаемую ошибку; значения не возвращает.
expect_failure() {
    expected=$1
    shift
    status=0
    output=$(./build/cmd-emulator "$@" 2>&1) || status=$?
    if [ "$status" -ne 1 ]; then
        printf 'Ожидался код 1, получен %s: %s\n' "$status" "$output" >&2
        exit 1
    fi
    case "$output" in
        *"Ошибка запуска:"*"$expected"*) ;;
        *) printf 'Нет ожидаемой ошибки %s: %s\n' "$expected" "$output" >&2; exit 1 ;;
    esac
    printf 'OK: %s\n' "$expected"
}

expect_failure 'укажите путь к VFS'
expect_failure 'неизвестный параметр' --unknown
expect_failure 'требуется значение' --vfs
expect_failure 'требуется значение' --prompt --vfs testdata/stage2-vfs.xml
expect_failure 'указан повторно' --vfs testdata/stage2-vfs.xml --vfs testdata/stage2-vfs.xml
expect_failure 'неверный JSON' --config tests/data/invalid-config.json
expect_failure 'открыть конфигурацию' --config tests/data/missing-config.json
expect_failure 'открыть VFS' --vfs tests/data/missing-vfs.xml
expect_failure 'открыть стартовый скрипт' --vfs testdata/stage2-vfs.xml --script tests/data/missing-script.txt
expect_failure 'обычным файлом' --vfs testdata
