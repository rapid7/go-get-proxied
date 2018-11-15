#!/bin/bash

if [ ! -z "$(gofmt -s -l $(find ./ -name '*.go' | grep -v vendor))" ] ; then
   gofmt -s -d $(find ./ -name '*.go' | grep -v vendor)
   echo "Go code is not formatted. Please run \"gofmt -s -w \$(find ./ -name '*.go' | grep -v vendor)\""
   exit 1
fi