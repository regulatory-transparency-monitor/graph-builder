package main

import (
	TMS "github.com/regulatory-transparency-monitor/graph-builder/internal"
)

func main() {

	transparencyMonitoringService := TMS.Init()
	transparencyMonitoringService.InitRoutes()
	transparencyMonitoringService.Run()

}
