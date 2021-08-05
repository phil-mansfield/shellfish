module github.com/phil-mansfield/shellfish

go 1.16

replace github.com/phil-mansfield/shellfish/cmd => ./cmd

replace github.com/phil-mansfield/shellfish/logging => ./logging

replace github.com/phil-mansfield/shellfish/version => ./version

replace github.com/phil-mansfield/shellfish/cmd/env => ./cmd/env

replace github.com/phil-mansfield/shellfish/io => ./io

replace github.com/phil-mansfield/shellfish/cosmo => ./cosmo

replace github.com/phil-mansfield/shellfish/parse => ./parse

replace github.com/phil-mansfield/shellfish/los => ./los

replace github.com/phil-mansfield/shellfish/los/tree => ./los/tree

replace github.com/phil-mansfield/shellfish/los/geom => ./los/geom

replace github.com/phil-mansfield/shellfish/los/analyze => ./los/analyze

replace github.com/phil-mansfield/shellfish/los/analyze/ellipse_grid => ./los/analyze/ellipse_grid

replace github.com/phil-mansfield/shellfish/los/analyze/math/calc => ./math/calc

replace github.com/phil-mansfield/shellfish/los/analyze/math/rand => ./math/rand

replace github.com/phil-mansfield/shellfish/los/analyze/math/sort => ./math/sort

replace github.com/phil-mansfield/shellfish/los/analyze/math/mat => ./math/mat

replace github.com/phil-mansfield/shellfish/los/analyze/math/interpolate => ./math/interpolate

replace github.com/phil-mansfield/shellfish/cmd/catalog => ./cmd/catalog

replace github.com/phil-mansfield/shellfish/cmd/halo => ./cmd/halo

replace github.com/phil-mansfield/shellfish/cmd/memo => ./cmd/memo

replace github.com/phil-mansfield/shellfish/math/rand => ./math/rand

replace github.com/phil-mansfield/shellfish/math/mat => ./math/mat

replace github.com/phil-mansfield/shellfish/math/sort => ./math/sort

replace github.com/phil-mansfield/shellfish/math/calc => ./math/calc

replace github.com/phil-mansfield/shellfish/math/interpolate => ./math/interpolate

require (
	github.com/phil-mansfield/shellfish/cmd v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/cmd/env v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/io v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/logging v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/version v0.0.0-00010101000000-000000000000
)
