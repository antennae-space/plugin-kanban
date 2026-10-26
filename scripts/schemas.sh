#!/bin/sh
# Writes the toolSchemas of plugin.json from the MCP tools of the backend (go run ./backend
# schemas): the IDE lists them from the manifest without starting the backend. CI checks that
# plugin.json is up to date.
set -eu
cd "$(dirname "$0")/.."
tmp=$(mktemp)
go run ./backend schemas > "$tmp"
node -e '
const fs = require("fs")
const m = JSON.parse(fs.readFileSync("plugin.json", "utf8"))
const schemas = JSON.parse(fs.readFileSync(process.argv[1], "utf8"))
m.contributes.mcpTools = Object.keys(schemas).sort()
m.contributes.toolSchemas = schemas
fs.writeFileSync("plugin.json", JSON.stringify(m, null, 2) + "\n")
' "$tmp"
rm -f "$tmp"
