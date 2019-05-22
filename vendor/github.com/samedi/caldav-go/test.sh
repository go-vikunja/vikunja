#!/usr/bin/env bash

go test -race ./...
rm -rf test-data
