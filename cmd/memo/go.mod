module github.com/phil-mansfield/shellfish/cmd/memo

go 1.16

replace github.com/phil-mansfield/shellfish/cmd/env => ../env

replace github.com/phil-mansfield/shellfish/cmd/halo => ../halo

replace github.com/phil-mansfield/shellfish/cmd/io => ../../io

replace github.com/phil-mansfield/shellfish/cmd/catalog => ../catalog

replace github.com/phil-mansfield/shellfish/io => ../../io

replace github.com/phil-mansfield/shellfish/cosmo => ../../cosmo

require (
	github.com/phil-mansfield/shellfish/cmd/env v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/cmd/halo v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/io v0.0.0-00010101000000-000000000000
)
