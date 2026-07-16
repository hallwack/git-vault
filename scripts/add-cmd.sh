#!/usr/bin/env bash
#
# scripts/add-cmd.sh
#
# Generates a new Cobra command and relocates it from cmd/ into
# internal/app/, renaming the package to match project structure
# (see architecture.md: CLI logic lives in internal/app, not cmd/).
#
# Usage:
#   ./scripts/add-cmd.sh <command-name>
#
# Example:
#   ./scripts/add-cmd.sh unlock
#   -> generates cmd/unlock.go
#   -> moves it to internal/app/unlock.go
#   -> renames "package cmd" to "package app"
#   -> registers it in internal/app/cli.go (reminder only, manual step)

set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <command-name>"
  exit 1
fi

CMD_NAME="$1"
CMD_FILE="cmd/${CMD_NAME}.go"
DEST_FILE="internal/app/${CMD_NAME}.go"

if [ ! -d "cmd" ] || [ ! -d "internal/app" ]; then
  echo "Error: run this from the project root (expects cmd/ and internal/app/ to exist)."
  exit 1
fi

if ! command -v cobra-cli &> /dev/null; then
  echo "Error: cobra-cli not found. Install it with:"
  echo "  go install github.com/spf13/cobra-cli@latest"
  exit 1
fi

echo "Generating command '${CMD_NAME}' with cobra-cli..."
cobra-cli add "${CMD_NAME}"

if [ ! -f "${CMD_FILE}" ]; then
  echo "Error: expected ${CMD_FILE} was not generated."
  exit 1
fi

echo "Moving ${CMD_FILE} -> ${DEST_FILE}..."
mv "${CMD_FILE}" "${DEST_FILE}"

echo "Renaming package cmd -> app in ${DEST_FILE}..."
sed -i.bak 's/^package cmd$/package app/' "${DEST_FILE}"
rm -f "${DEST_FILE}.bak"

echo ""
echo "Done. Created ${DEST_FILE}."
echo ""
echo "Next steps:"
echo "  1. Open ${DEST_FILE} and implement the command logic."
echo "  2. Register it in internal/app/cli.go:"
echo "       rootCmd.AddCommand(New$(echo "${CMD_NAME:0:1}" | tr '[:lower:]' '[:upper:]')${CMD_NAME:1}Cmd())"
echo "  3. Run: go run ./cmd/git-vault ${CMD_NAME}"
