#!/usr/bin/env bash

MONEX_HOME="/apps/math"
cd $MONEX_HOME

echo "$MONEX_HOME"
echo "Fixing session file permissions..."

# Fix sessionfile permissions
# chmod 775 $MONEX_HOME/data/*.datc

echo "Session file permissions fixed!"