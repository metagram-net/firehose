package firehose

import (
	"github.com/metagram-net/firehose/apigen/gen"
	"github.com/metagram-net/firehose/apigen/schema"
)

Server: gen.#Server & {
	@gen(server)
	Schema: _Server

	Outdir: "."

	PartialDirs: ["./apigen/partials"]
	TemplateDirs: ["./apigen/templates"]

	// CUE needs to resolve everything to a concrete value. Use the empty
	// string here to tell Hof that this is the top-level module.
	PackageName: ""
}

_Server: schema.#Server & {
	GoModule:  "github.com/metagram-net/firehose"
	GoPackage: "server"

	Groups: _groups
}

_groups: [
	for g, routes in _group {
		{
			Name: g
			Routes: [
				for r in routes {
					r & {Group: g}
				},
			]
		}
	},
]

#GET: schema.#Route & {
	Method:  "GET"
	Path:    string
	Params?: string
	// By standard, a GET request can't have a body.
}

#POST: schema.#Route & {
	Method: "POST"
	Body:   string
	// By convention, a POST request can't have path params.
}

_group: [string]: schema.#Routes

_group: "WellKnown": [
	#GET & {
		Name:          "HealthCheck"
		Authenticated: false
		Path:          "/.well-known/health-check"
		Return:        "wellknown.HealthCheckResponse"
	},
]

_group: "Auth": [
	#GET & {
		Name:   "Whoami"
		Path:   "/auth/whoami"
		Return: "user.User"
	},
]

_group: "Drops": [
	#GET & {
		Name:   "Next"
		Path:   "/v1/drops/next"
		Return: "drop.Drop"
	},
	#GET & {
		Name:   "Get"
		Path:   "/v1/drops/get/{id}"
		Params: "drop.GetParams"
		Return: "drop.Drop"
	},
	#POST & {
		Name:   "List"
		Path:   "/v1/drops/list"
		Body:   "drop.ListBody"
		Return: "drop.ListResponse"
	},
	#POST & {
		Name:   "Create"
		Path:   "/v1/drops/create"
		Body:   "drop.CreateBody"
		Return: "drop.Drop"
	},
	#POST & {
		Name:   "Update"
		Path:   "/v1/drops/update"
		Body:   "drop.UpdateBody"
		Return: "drop.Drop"
	},
	#POST & {
		Name:   "Move"
		Path:   "/v1/drops/move"
		Body:   "drop.MoveBody"
		Return: "drop.Drop"
	},
	#POST & {
		Name:   "Delete"
		Path:   "/v1/drops/delete"
		Body:   "drop.DeleteBody"
		Return: "drop.Drop"
	},
]
