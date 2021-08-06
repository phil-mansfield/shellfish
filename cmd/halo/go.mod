module github.com/phil-mansfield/shellfish/cmd/halo

go 1.16

replace github.com/phil-mansfield/shellfish/cmd/catalog => ../catalog

replace github.com/phil-mansfield/shellfish/cosmo => ../../cosmo

replace github.com/phil-mansfield/shellfish/io => ../../io

require (
	github.com/phil-mansfield/shellfish/cmd/catalog v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/cosmo v0.0.0-00010101000000-000000000000
	github.com/phil-mansfield/shellfish/io v0.0.0-00010101000000-000000000000
)
