package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/regulatory-transparency-monitor/graph-builder/config"
	"github.com/regulatory-transparency-monitor/graph-builder/graph"
	"github.com/regulatory-transparency-monitor/graph-builder/graph/generated"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/manager"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/repository"
	service "github.com/regulatory-transparency-monitor/graph-builder/internal/service"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"

	"github.com/spf13/viper"
)

// App main application object
type App struct {
	Router  *mux.Router
	Service *service.Service
	Manager *manager.Manager
}

// Init initializes app
func Init() *App {
	// 1) Read configurations
	err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Loading config failure: ", err)
	}

	// 2) Connect to Neo4j
	neo4Conn, err := repository.NewNeo4jConnection()
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	r := &repository.Neo4jRepository{
		Connection: neo4Conn,
	}

	// 3) Instantiate Service
	srv := service.NewService(r)

	// 4) Instantiate mngrestrator
	tf := dataparser.TransformerRegistry
	mngr := manager.NewManager(tf, srv)
	err = mngr.Start()
	if err != nil {
		logger.Error("mngrestrator failure: ", err)
	}

	return &App{
		Service: srv,
		Manager: mngr,
	}
}

// Run executes app
func (a *App) Run() {
	host := viper.GetString("SERVER_IP")
	port := viper.GetString("API_PORT")
	addrStr := fmt.Sprintf("%s:%s", host, port)
	logger.Info("GraphQL API Listening: ", logger.LogFields{"api_url": addrStr})
	logger.Fatal("Service failure", http.ListenAndServe(addrStr, a.Router))
}

// InitRoutes initializing all the routes
func (a *App) InitRoutes() {
	a.Router = mux.NewRouter()
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			Service: a.Service}}))
	a.Router.Handle("/playground", playground.Handler("GoNeo4jGql GraphQL playground", "/instance"))
	a.Router.Handle("/instance", srv)
}
