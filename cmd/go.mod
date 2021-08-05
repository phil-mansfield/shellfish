module github.com/phil-mansfield/shellfish/cmd

go 1.16

replace github.com/phil-mansfield/shellfish/cmd/catalog => ./catalog

replace github.com/phil-mansfield/shellfish/cmd/env => ./env

replace github.com/phil-mansfield/shellfish/cmd/halo => ./halo

replace github.com/phil-mansfield/shellfish/cmd/memo => ./memo

replace github.com/phil-mansfield/shellfish/cosmo => ../cosmo

replace github.com/phil-mansfield/shellfish/io => ../io

replace github.com/phil-mansfield/shellfish/logging => ../logging

replace github.com/phil-mansfield/shellfish/los => ../los

replace github.com/phil-mansfield/shellfish/analyze => ../analyze

replace github.com/phil-mansfield/shellfish/los/analyze => ../los/analyze

replace github.com/phil-mansfield/shellfish/los/geom => ../los/geom

replace github.com/phil-mansfield/shellfish/los/tree => ../los/tree

replace github.com/phil-mansfield/shellfish/math/rand => ../math/rand

replace github.com/phil-mansfield/shellfish/math/sort => ../math/sort

replace github.com/phil-mansfield/shellfish/parse => ../parse

replace github.com/phil-mansfield/shellfish/version => ../version

replace github.com/phil-mansfield/shellfish/math/mat => ../math/mat

replace github.com/phil-mansfield/shellfish/analyze/ellipse_grid => ../los/analyze/ellipse_grid

replace github.com/phil-mansfield/shellfish/los/analyze/ellipse_grid => ../los/analyze/ellipse_grid

replace github.com/phil-mansfield/shellfish/math/interpolate => ../math/interpolate

require (
	github.com/phil-mansfield/shellfish/cmd/catalog v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/cmd/env v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/cmd/halo v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/cmd/memo v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/cosmo v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/io v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/logging v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/los v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/los/analyze v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/los/geom v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/los/tree v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/math/rand v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/math/sort v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/parse v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/version v0.0.0-00010101000000-000000000000
)
