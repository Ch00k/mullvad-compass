#!/bin/bash
set -e

# Check for uncommitted changes first
if ! git diff-index --quiet HEAD --; then
    echo "Error: You have uncommitted changes. Please commit or stash them first."
    exit 1
fi

# Fetch latest remote state and pull changes
echo "Fetching latest remote tags and changes..."
git fetch --tags origin
git pull origin main

# Get latest tag from git
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
echo "Latest tag: $LATEST_TAG"

# Extract version numbers
IFS='.' read -r MAJOR MINOR PATCH <<<"$LATEST_TAG"

# Default to patch release
RELEASE_TYPE=${1:-patch}

case $RELEASE_TYPE in
major)
    NEW_VERSION="$((MAJOR + 1)).0.0"
    ;;
minor)
    NEW_VERSION="$MAJOR.$((MINOR + 1)).0"
    ;;
patch)
    NEW_VERSION="$MAJOR.$MINOR.$((PATCH + 1))"
    ;;
*)
    # Custom version provided
    if [[ $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        NEW_VERSION="$1"
    else
        echo "Usage: $0 [major|minor|patch|1.2.3]"
        echo "Examples:"
        echo "  $0 patch    # 1.0.0 -> 1.0.1"
        echo "  $0 minor    # 1.0.0 -> 1.1.0"
        echo "  $0 major    # 1.0.0 -> 2.0.0"
        echo "  $0 2.1.0   # specific version"
        exit 1
    fi
    ;;
esac

echo "New version: $NEW_VERSION"

read -p "Create release $NEW_VERSION? [y/N] " -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Writing version to .version file..."
    echo "$NEW_VERSION" >.version

    echo "Updating README with new version..."
    ./update-readme-usage.sh

    # Check if README or .version were modified
    if ! git diff --quiet README.md .version; then
        echo "Committing README and .version changes..."
        git add README.md .version
        git commit -m "Update README and version to $NEW_VERSION"
    else
        echo "No changes to commit"
    fi

    echo "Creating annotated tag..."
    git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"

    echo "Pushing to origin..."
    git push origin main "$NEW_VERSION"

    echo "Release $NEW_VERSION created successfully!"
    echo "Binary build will start automatically via GitHub Actions."
    echo "Monitor the GitHub Actions page for build progress."
else
    echo "Release cancelled"
fi
