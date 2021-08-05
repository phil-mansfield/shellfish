module github.com/phil-mansfield/shellfish/los/analyze

go 1.16

replace github.com/phil-mansfield/shellfish/math/interpolate => ../../math/interpolate

replace github.com/phil-mansfield/shellfish/los => ../../los

replace github.com/phil-mansfield/shellfish/math/mat => ../../math/mat

replace github.com/phil-mansfield/shellfish/los/geom => ../geom

replace github.com/phil-mansfield/shellfish/math/sort => ../../math/sort

replace github.com/phil-mansfield/shellfish/los/analyze/ellipse_grid => ./ellipse_grid

replace github.com/phil-mansfield/shellfish/io => ../../io

replace github.com/phil-mansfield/shellfish/cosmo => ../../cosmo

require (
	github.com/gonum/blas v0.0.0-20181208220705-f22b278b28ac // indirect
	github.com/gonum/floats v0.0.0-20181209220543-c233463c7e82 // indirect
	github.com/gonum/internal v0.0.0-20181124074243-f884aa714029 // indirect
	github.com/gonum/lapack v0.0.0-20181123203213-e4cdc5a0bff9 // indirect
	github.com/gonum/matrix v0.0.0-20181209220409-c518dec07be9
	github.com/phil-mansfield/pyplot v0.1.0
	github.com/phil-mansfield/shellfish/los v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/los/analyze/ellipse_grid v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/los/geom v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/math/interpolate v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/math/mat v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/math/sort v0.0.0-00010101000000-000000000000
)
