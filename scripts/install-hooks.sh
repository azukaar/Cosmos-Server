#!/bin/bash
cd "$(dirname "$0")/.."
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
echo "Git hooks installed."
