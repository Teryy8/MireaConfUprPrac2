#!/usr/bin/env sh
# Сравнение трёх SVG с деревом реального npm во временных папках.
set -eu
cd "$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
command -v npm >/dev/null 2>&1 || { echo "Для сравнения нужен npm" >&2; exit 1; }
comparison_dir=$(mktemp -d)
trap 'rm -rf "$comparison_dir"' EXIT HUP INT TERM
for package in is-odd is-even supports-color; do
    case "$package" in
        is-odd) version=3.0.1 ;;
        is-even) version=1.0.0 ;;
        supports-color) version=5.5.0 ;;
    esac
    ./run.sh --config "configs/visual-$package.json"
    package_dir="$comparison_dir/$package"
    mkdir "$package_dir"
    printf '{"name":"comparison","version":"1.0.0","private":true,"dependencies":{"%s":"%s"}}\n' \
        "$package" "$version" > "$package_dir/package.json"
    npm install --prefix "$package_dir" --ignore-scripts --no-audit --no-fund
    npm ls --prefix "$package_dir" --all
    echo "Сравните дерево с файлом $package.svg"
done
