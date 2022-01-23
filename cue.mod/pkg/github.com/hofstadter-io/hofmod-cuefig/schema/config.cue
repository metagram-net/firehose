package schema

import (
  "strings"
)

// A general configuration 
#Config: {
  // binary name, often last part of Package
  Name: string
  configName: strings.ToCamel(Name)
  ConfigName: strings.ToTitle(Name)

	// Config location type
	Location: *"local" | "user" | "cache" | "system"
	// Where the cue context starts
	Workpath: string | *""
	// Entrypoint from the Workpath
  Entrypoint: string | *"\(configName).cue"

	// Defaults to use when loading
	Defaults: {
		Expression: string | *""
		Package: string | *""
		LabelExprs: [...string] | *[]
	}

	// Default to sensative content
	Sensative: bool | *true

	// User provided config schema
  ConfigSchema: {...}
}
