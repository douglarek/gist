package main

import "flag"

var (
	list, all         bool
	get               string
	login             bool
	delete            StringSliceValue
	description       string
	anonymous, public bool
)

func init() {
	flag.BoolVar(&list, "l", false, "List public gists, with -A list all ones")
	flag.BoolVar(&all, "A", false, "")
	flag.StringVar(&get, "i", "", "Get a gist by id")
	flag.BoolVar(&login, "login", false, "Authenticate gist on this computer")
	flag.Var(&delete, "D", "Detele existing gists by ids")
	flag.StringVar(&description, "d", "", "Adds a description to your gist")
	flag.BoolVar(&anonymous, "a", false, "Create an anonymous gist")
	flag.BoolVar(&public, "p", false, "Makes your gist public")
}

func main() {
	flag.Parse()

	g := &Gist{client}
	switch {
	case flag.NArg() != 0:
		exit(g.Create(description, anonymous, public, flag.Args()...))
	case len(get) != 0:
		exit(g.Get(get))
	case login:
		exit(token(ask()))
	case list:
		exit(g.List(!all))
	case len(delete) != 0:
		exit(g.Delete(delete...))
	default:
		flag.Usage()
	}
}
