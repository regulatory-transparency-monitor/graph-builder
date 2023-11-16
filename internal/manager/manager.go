package manager

import (
	"fmt"
	"time"

	services "github.com/regulatory-transparency-monitor/graph-builder/internal/service"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/versioning"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/plugin"
)

type Manager struct {
	Transformers   map[string]dataparser.Transformer
	Service        *services.Service
	VersionManager *versioning.VersionManager
	Scheduler      *Scheduler
	PluginManager  *plugin.PluginManager
}

func NewManager(tf map[string]dataparser.Transformer, srv *services.Service) *Manager {
	version, err := srv.GetLatestVersion()
	if err != nil {
		logger.Warning("Couldn't fetch latest version, initializing with version 0.0.1")
		version = "0.0.1"
	}

	vm := versioning.NewVersionManager(version)
	pluginMgr := plugin.NewPluginManager()
	pluginMgr.RegisterPluginConstructors()
	pluginMgr.InitializePlugins()

	o := &Manager{
		Transformers:   tf,
		Service:        srv,
		VersionManager: vm,
		Scheduler:      NewScheduler(),
		PluginManager:  pluginMgr,
	}

	return o
}

func (o *Manager) Start() error {
	// 1) Create the initial version node
	err := o.Service.SetupUUIDForKnownLabels()
	if err != nil {
		logger.Error("Failed to create UUID constraints: %v", err)
		return err
	}
	// 2) Run Initial infrastructure scan
	err = o.coordinator()
	if err != nil {
		return err
	}
	// 3) Start periodic scans
	o.startPeriodicScans()

	return nil
}

// startPeriodicScans scans the infrastructure periodically using provider plugins
func (o *Manager) startPeriodicScans() {
	o.Scheduler.AddTask("@every 3m", func() {
		o.VersionManager.IncrementVersion()
		o.coordinator()
	})
	o.Scheduler.Start()
}

func getCurrentTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (o *Manager) coordinator() error {

	v := o.VersionManager.GetCurrentVersion()

	err := o.Service.CreateMetadataNode(v, getCurrentTimeString())
	if err != nil {
		logger.Error("Failed to create metadata node: %v", err)
		return err
	}
	logger.Info("*** Start fetching resources *** ")
	// 0) Fetch API services using the appropriate plugin
	for providerType := range o.PluginManager.ActivePlugins {
		logger.Info("Fetching API services using ", logger.LogFields{"provider plugin": providerType})
		// 1) Scan infrastructure using the appropriate plugin
		rawDataMap := plugin.Scanner(o.PluginManager, providerType)

		// 2) Transform raw data into generic data using the appropriate transformer
		genericData, err := dataparser.TransformData(rawDataMap) // Call the TransformData function here
		if err != nil {
			logger.Error("Error transforming data: %v", err)
			continue // continue to next provider if there's an error
		}
		logger.Info("*** Generic data transformed ***")

		// 3) Store generic data in Neo4j
		for _, component := range genericData {

			uuid, err := o.Service.CreateInfrastructureComponent(v, component)
			if err != nil {
				logger.Error(fmt.Sprintf("Error storing %s in Neo4j: %v", component.Type, err))
				continue
			}

			if component.Type == "Project" {
				err = o.Service.LinkProjectToMetadata(v, uuid)
				if err != nil {
					logger.Error("Failed to link project to metadata: %v", err)
				}
			}
		}

		// 4) Relationship Creation Phase: Create relationships between nodes
		for _, component := range genericData {
			err := o.Service.CreateRelationships(v, component)
			if err != nil {
				logger.Error(fmt.Sprintf("Error creating Relationship %s in Neo4j: %v", component.Type, err))
				continue
			}
		}

	}
	logger.Info("*** Finsihed storing data for all plugins ***")

	return nil
}
