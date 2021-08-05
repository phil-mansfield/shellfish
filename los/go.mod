module github.com/phil-mansfield/shellfish/los

go 1.16

replace github.com/phil-mansfield/shellfish/los/geom => ./geom

replace github.com/phil-mansfield/shellfish/math/mat => ../math/mat

replace github.com/phil-mansfield/shellfish/io => ../io

replace github.com/phil-mansfield/shellfish/math/sort => ../math/sort

require (
	github.com/phil-mansfield/shellfish/io v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/los/geom v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/math/mat v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/math/sort v0.0.0-00010101000000-000000000000
)

replace github.com/phil-mansfield/shellfish/cosmo => ../cosmo
