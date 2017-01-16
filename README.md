# rdl-gen-ell-server

A simple generator to produce a server written in [ell](https://github.com/boynton/ell) from 
an API spec written in [RDL](https://github.com/ardielle).

## Installation

Make sure Ell is installed/current:

	$ go get -u github.com/boynton/ell/...

Make sure RDL itself is installed:

	$ go get -u github.com/ardielle/ardielle-tools/...

Then install this generator:

	$ go get github.com/boynton/rdl-gen-ell-server

## Usage

An example RDL file is included [here](https://github.com/boynton/rdl-gen-ell-server/blob/master/example.rdl). To 
generate the scaffolding for an HTTP server from this:

	$ rdl generate ell-server example.rdl

This produces a file `example.ell`. To run the server:

	$ ell example.rdl
	[web server running at http://localhost:8080]

Try hitting it with curl:

	$ curl http://localhost:8080/book
	{"status": 501, "data": "Not Implemented"}

Replaced the stub implementations with your own. This example matches the http-example.ell file in the
ell release, where a functional implementation exists.
