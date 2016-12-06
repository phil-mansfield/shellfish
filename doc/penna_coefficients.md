# Penna-Dines Coefficients

Coming soon!

Penna-Dines functions are a type of basis function which is particularly good at 
representing 3D almost-ellisoidal blobby shapes (much like spherical harmonics).
They were originally designed to characterize the shape of tumors and are therefore
good at representing lobed shapes, even at a low order. The exact functional form is:
```
r(phi, theta) = sum_{i=0}^I sum_{j=0}^J sum_{k=0}^K c_{ijk} * sin(theta)^(i+j) * cos(theta)^k
                                                            * sin(phi)^j * cos(phi)^i
```
Here, the convention is that phi is the azimuthal angle and theta is the polar angle.
Shellfish sets K=1 and I=J=P-1, where P is the "order" of the Penna-Dines function, meaning
that at order P there are 2P^2 `c_{ijk}` values.

A better formatted version of this equation can be found in equation 4 of our paper
(or in the original [Penna-Dines](https://www.computer.org/csdl/trans/tp/2007/09/i1673.pdf)
paper.) An example of a function which evaluates a set of Penna-Dines coordinates can
be found [here](https://github.com/phil-mansfield/shellfish/blob/master/los/analyze/penna.go#L62)
and can easily be translated into the programming language of your choice. The array
`cs` contains `c_{ijk}` values in the same order as the output of `shellfish shell`.
