#!/bin/bash 

go get -u github.com/gonum/matrix
go get -u github.com/gonum/floats
go get -u github.com/gonum/internal/asm/c64
go get -u github.com/gonum/internal/asm/c128
go get -u github.com/gonum/internal/asm/f32
go get -u github.com/gonum/internal/asm/f64
go get -u github.com/phil-mansfield/consistent_trees
go get -u github.com/phil-mansfield/go-artio

cd $GOPATH/src/github.com/phil-mansfield/shellfish && git pull && go install && cd -
