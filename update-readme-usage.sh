#!/usr/bin/env bash
set -e

# Accept version as parameter or fetch from GitHub
VERSION="$1"

# Build the project
echo "Building project..."
make build

if [ -z "$VERSION" ]; then
    # Get the latest version from GitHub if not provided
    echo "Fetching latest version from GitHub..."
    VERSION=$(curl -s https://api.github.com/repos/Ch00k/mullvad-compass/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "Warning: Could not fetch latest version from GitHub, using '0.0.1'"
        VERSION="0.0.1"
    else
        echo "Latest version: $VERSION"
    fi
else
    echo "Using provided version: $VERSION"
fi

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
