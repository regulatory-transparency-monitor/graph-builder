package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/regulatory-transparency-monitor/graph-builder/graph"
	"github.com/regulatory-transparency-monitor/graph-builder/graph/generated"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/db"
	service "github.com/regulatory-transparency-monitor/graph-builder/internal/service"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/spf13/viper"
)

// App main application object
type App struct {
	Router  *mux.Router
	Service service.Service
}

// Init initializes app
func Init() *App {
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("API_PORT", "8080")
	viper.SetDefault("LOGGER_FORMATTER", "console")
	viper.SetDefault("LOGGER_LEVEL", "debug")
	viper.SetDefault("NEO4J_HOST", "localhost")
	viper.SetDefault("NEO4J_PORT", "7687")
	viper.SetDefault("NEO4J_USER", "neo4j")
	viper.SetDefault("NEO4J_PASS", "testingshit")
	viper.SetDefault("NEO4J_PROTO", "bolt")

	neo4Conn, err := db.NewNeo4jConnection()
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	r := &db.Neo4jRepository{
		Connection: neo4Conn,
	}

	return &App{
		Service: service.NewService(r),
	}

}

// Run executes app
func (a *App) Run() {
	//addrStr := fmt.Sprintf("localhost:%s", viper.Get("API_PORT"))
	//logger.Info("API Listening", logger.LogFields{"api_url": addrStr})
	//logger.Fatal("Service failure", http.ListenAndServe(addrStr, a.Router))


	host := "192.168.2.139" // TODO local ip so far
    port := viper.GetString("API_PORT")
    addrStr := fmt.Sprintf("%s:%s", host, port)
    logger.Info("API Listening", logger.LogFields{"api_url": addrStr})
    logger.Fatal("Service failure", http.ListenAndServe(addrStr, a.Router))


}

// InitRoutes initializing all the routes
func (a *App) InitRoutes() {
	a.Router = mux.NewRouter()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Service: a.Service}}))
	a.Router.Handle("/playground", playground.Handler("GoNeo4jGql GraphQL playground", "/query"))
	a.Router.Handle("/query", srv)
}
