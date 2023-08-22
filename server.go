package main

import (
	tms "github.com/regulatory-transparency-monitor/graph-builder/internal"
)

func main() {
	transparencyMonitoringService := tms.Init()
	transparencyMonitoringService.InitRoutes()
	transparencyMonitoringService.Run()
}
