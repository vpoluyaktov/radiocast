#!/bin/bash

# Script to bump semantic version
# Usage: ./bump-version.sh [major|minor|patch|stage]

set -e

VERSION_FILE="VERSION"
BUMP_TYPE=${1:-patch}

if [ ! -f "$VERSION_FILE" ]; then
    echo "Error: VERSION file not found"
    exit 1
fi

CURRENT_VERSION=$(cat $VERSION_FILE)
echo "Current version: $CURRENT_VERSION"

# Check if current version has pre-release suffix
if [[ $CURRENT_VERSION == *"-rc."* ]]; then
    # Extract base version and RC number
    BASE_VERSION=$(echo $CURRENT_VERSION | cut -d'-' -f1)
    RC_PART=$(echo $CURRENT_VERSION | cut -d'-' -f2)
    RC_NUMBER=$(echo $RC_PART | cut -d'.' -f2)
    
    IFS='.' read -r -a VERSION_PARTS <<< "$BASE_VERSION"
    MAJOR=${VERSION_PARTS[0]}
    MINOR=${VERSION_PARTS[1]}
    PATCH=${VERSION_PARTS[2]}
else
    # Regular version without pre-release
    IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
    MAJOR=${VERSION_PARTS[0]}
    MINOR=${VERSION_PARTS[1]}
    PATCH=${VERSION_PARTS[2]}
    RC_NUMBER=0
fi

# Bump version based on type
case $BUMP_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        NEW_VERSION="$MAJOR.$MINOR.$PATCH"
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        NEW_VERSION="$MAJOR.$MINOR.$PATCH"
        ;;
    patch)
        PATCH=$((PATCH + 1))
        NEW_VERSION="$MAJOR.$MINOR.$PATCH"
        ;;
    stage)
        # For staging, create or increment RC version
        if [[ $CURRENT_VERSION == *"-rc."* ]]; then
            # Increment RC number
            RC_NUMBER=$((RC_NUMBER + 1))
            NEW_VERSION="$MAJOR.$MINOR.$PATCH-rc.$RC_NUMBER"
        else
            # Create first RC version
            NEW_VERSION="$MAJOR.$MINOR.$PATCH-rc.1"
        fi
        ;;
    *)
        echo "Error: Invalid bump type. Use major, minor, patch, or stage"
        exit 1
        ;;
esac

echo "New version: $NEW_VERSION"

# Update VERSION file
echo "$NEW_VERSION" > $VERSION_FILE
echo "Version updated to $NEW_VERSION"
