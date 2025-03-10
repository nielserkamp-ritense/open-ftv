#!/bin/sh
TAB=0
while read line; do
  if [ "replace (" = "$line" ]; then
    break
  fi
  if [ ")" = "$line" ]; then
    TAB=0
  fi
  if [ "$TAB" -eq "1" ]; then
    echo -e "\t$line"
  else
    echo "$line"
  fi
  if [ "require (" = "$line" ]; then
    TAB=1
  fi
done <"$1"
