package gen

import (
	"strings"

	"github.com/hofstadter-io/hof/schema/gen"

	"github.com/metagram-net/firehose/apigen/schema"
)

#Client: gen.#HofGenerator & {
	Schema: schema.#Client

	// Outdir is the base directory for all generated files.
	Outdir: string

	// In is generator context passed to every template.
	In: {
		Client: Schema
	}

	// Out describes the output files to generate.
	Out: [...gen.#HofGeneratorFile] & [
		{
			TemplatePath: "client/endpoints.go.tmpl"
			Filepath:     "\(Outdir)/client/endpoints.apigen.go"
		},
	]

	PartialDirs: [...string] | *[]
	Partials:    [...gen.#Templates] & [

		for _, path in PartialDirs {
			_path: strings.TrimSuffix(path, "/")
			Globs: [_path + "/**/*"]
			TrimPrefix: _path + "/"
		},
	]

	TemplateDirs: [...string] | *[]
	Templates:    [...gen.#Templates] & [

		for _, path in TemplateDirs {
			_path: strings.TrimSuffix(path, "/")
			Globs: [_path + "/**/*"]
			TrimPrefix: _path + "/"
		},
	]
}
