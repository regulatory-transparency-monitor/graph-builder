package graph

import service "github.com/regulatory-transparency-monitor/graph-builder/internal/service"

//go:generate go run github.com/99designs/gqlgen generate
// This file will not be regenerated automatically.
//
// It serves as dependency injection for the app, add any dependencies required here.

type Resolver struct {
	Service *service.Service
}
