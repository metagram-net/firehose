package firehose

import (
	"github.com/metagram-net/firehose/apigen/gen"
	"github.com/metagram-net/firehose/apigen/schema"
)

_GoModule: "github.com/metagram-net/firehose"

Server: gen.#Server & {
	@gen(server)
	Schema: schema.#Server & {
		GoModule:  _GoModule
		GoPackage: "server"
		API:       _API
	}

	Outdir: "."

	PartialDirs: ["./apigen/partials"]
	TemplateDirs: ["./apigen/templates"]

	// CUE needs to resolve everything to a concrete value. Use the empty
	// string here to tell Hof that this is the top-level module.
	PackageName: ""
}

Client: gen.#Client & {
	@gen(client)
	Schema: schema.#Client & {
		GoModule:  _GoModule
		GoPackage: "client"
		API:       _API
	}

	Outdir: "."

	PartialDirs: ["./apigen/partials"]
	TemplateDirs: ["./apigen/templates"]

	// CUE needs to resolve everything to a concrete value. Use the empty
	// string here to tell Hof that this is the top-level module.
	PackageName: ""
}
