package schema

#API: {
	Groups: #Groups
}

#HttpMethod: "GET" | "POST"

#Groups: [...#Group] | *[]
#Group: {
	Name:   string
	Routes: #Routes
}

#Routes: [...#Route] | *[]
#Route: {
	Name:  string
	Group: string

	Authenticated: bool | *true

	Path:    string
	Method:  #HttpMethod
	Params?: string
	Body?:   string

	Return: string
}
