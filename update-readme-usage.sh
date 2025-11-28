#!/usr/bin/env bash
set -e

# Read version from .version file
if [ -f ".version" ]; then
    VERSION=$(cat .version)
    echo "Using version from .version file: $VERSION"
else
    echo "Error: .version file not found"
    exit 1
fi

# Build the project
echo "Building project..."
make build

# Capture outputs and remove trailing whitespace
echo "Generating best-server output..."
BEST_SERVER_OUTPUT=$(./dist/mullvad-compass --deterministic-output | sed 's/[[:space:]]*$//')

echo "Generating multiple-servers output..."
MULTIPLE_SERVERS_OUTPUT=$(./dist/mullvad-compass --deterministic-output --max-distance 250 | sed 's/[[:space:]]*$//')

echo "Generating help output..."
HELP_OUTPUT=$(./dist/mullvad-compass --help | sed "s/ dev$/ $VERSION/" | sed 's/[[:space:]]*$//')

# Update README.md in-place
echo "Updating README.md..."
README_FILE="README.md"

# Create replacement texts with command prompts
BEST_SERVER_REPLACEMENT="<!-- best-server:start -->\n\`\`\`\n\\\$ mullvad-compass\n${BEST_SERVER_OUTPUT}\n\`\`\`\n<!-- best-server:end -->"
MULTIPLE_SERVERS_REPLACEMENT="<!-- multiple-servers:start -->\n\`\`\`\n\\\$ mullvad-compass --max-distance 250\n${MULTIPLE_SERVERS_OUTPUT}\n\`\`\`\n<!-- multiple-servers:end -->"
HELP_REPLACEMENT="<!-- help:start -->\n\`\`\`\n\\\$ mullvad-compass --help\n${HELP_OUTPUT}\n\`\`\`\n<!-- help:end -->"

# Use perl for in-place replacement
perl -i -0pe "s/<!-- best-server:start -->.*?<!-- best-server:end -->/${BEST_SERVER_REPLACEMENT}/s" "$README_FILE"
perl -i -0pe "s/<!-- multiple-servers:start -->.*?<!-- multiple-servers:end -->/${MULTIPLE_SERVERS_REPLACEMENT}/s" "$README_FILE"
perl -i -0pe "s/<!-- help:start -->.*?<!-- help:end -->/${HELP_REPLACEMENT}/s" "$README_FILE"

echo "README.md updated successfully"
