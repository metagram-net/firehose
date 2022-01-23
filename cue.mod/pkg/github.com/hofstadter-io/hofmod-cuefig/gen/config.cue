package gen

import (

  hof "github.com/hofstadter-io/hof/schema/gen"

  "github.com/hofstadter-io/hofmod-cuefig/schema"
)

#HofGenerator: hof.#HofGenerator & {
  Config: schema.#Config
  Outdir?: string

  In: {
    CONFIG: Config
  }

	Out: [...hof.#HofGeneratorFile] & [
		for _, F in OnceFiles { F },
	]

  PackageName: "github.com/hofstadter-io/hofmod-cuefig"

	Partials: []

  // OnceFiles that are not repeatedly used, they are generated once for the whole generator input
  OnceFiles: [...hof.#HofGeneratorFile] & [
    {
      TemplatePath:  "config.go"
      Filepath:  "\(Outdir)/cuefig/\(In.CONFIG.Name).go"
    },
  ]

}
