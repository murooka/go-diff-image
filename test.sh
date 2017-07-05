#!/bin/bash

for dir in testdata/*; do
  echo $dir
  diff-image -o=$dir/diff.png $dir/before.png $dir/after.png
done
