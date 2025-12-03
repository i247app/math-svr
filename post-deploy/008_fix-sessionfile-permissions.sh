#!/usr/bin/env bash

MATH_HOME="/apps/math"
cd $MATH_HOME

echo "$MATH_HOME"
echo "Fixing session file permissions..."

# Fix sessionfile permissions
# chmod 775 $MATH_HOME/data/*.datc

echo "Session file permissions fixed!"