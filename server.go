package main

import (
	app "github.com/regulatory-transparency-monitor/graph-builder/internal"
)

const defaultPort = "8080"

func main() {

	myApp := app.Init()
	myApp.InitRoutes()
	myApp.Run()
	//port := os.Getenv("PORT")
	/*if port == "" {
		port = defaultPort
	}

	 srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers:  &generated.Resolver{Service: a.Service},
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	})) 

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	//http.Handle("/query", srv)

	logger.Info("connect to http://localhost:%s/ for GraphQL playground", port)
	logger.Fatal(http.ListenAndServe(":"+port, nil))*/

}
