#!/bin/bash

if [ ! -z "$(gofmt -l $(find ./ -name '*.go' | grep -v vendor))" ] ; then
   gofmt -d $(find ./ -name '*.go' | grep -v vendor)
   echo "Go code is not formatted. Please run \"gofmt -w \$(find ./ -name '*.go' | grep -v vendor)\""
   exit 1
fi