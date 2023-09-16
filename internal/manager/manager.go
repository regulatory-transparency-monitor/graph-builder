package manager

import (
	"fmt"
	"time"

	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/plugin"
	services "github.com/regulatory-transparency-monitor/graph-builder/internal/service"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/versioning"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
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
	err = o.getInfrastructure()
	if err != nil {
		return err
	}
	// 3) Start periodic scans
	o.startPeriodicScans()

	return nil
}

// PeriodicScan scans the infrastructure periodically using provider plugins
func (o *Manager) startPeriodicScans() {
	o.Scheduler.AddTask("@every 30s", func() {
		o.VersionManager.IncrementVersion()
		o.getInfrastructure()
	})
	o.Scheduler.Start()
}

func getCurrentTimeString() string {
	// Assuming you want a specific time format, like "2006-01-02 15:04:05"
	return time.Now().Format("2006-01-02 15:04:05")
}

func (o *Manager) getInfrastructure() error {
	v := o.VersionManager.GetCurrentVersion()

	err := o.Service.CreateMetadataNode(v, getCurrentTimeString())
	if err != nil {
		logger.Error("Failed to create metadata node: %v", err)
		return err
	}

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
		logger.Info("Generic data transformed")

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

		// 4) Create relationships between nodes
		// 2. Relationship Creation Phase
		for _, component := range genericData {
			if component.Type == "Instance" {
				err := o.Service.CreateInstanceRelationships(component.ID, v, component.Relationships)
				if err != nil {
					logger.Error("Failed to create relationships for instance: %v", err)
				}
			}
			if component.Type == "ClusterNode" {
				err := o.Service.CreateClusterNodeRel(component.ID, v, component.Relationships)
				if err != nil {
					logger.Error("Failed to create relationships for cluster node: %v", err)
				}
			}
			if component.Type == "Pod" {
				err := o.Service.CreatePodRel(component.ID, v, component.Relationships)
				if err != nil {
					logger.Error("Failed to create relationships for pod: %v", err)
				}
			}
			if component.Type == "Volume" {
				err := o.Service.CreateVolumeRel(component.ID, v, component.Relationships)
				if err != nil {
					logger.Error("Failed to create relationships for volume: %v", err)
				}
			}

		}

	}
	logger.Info("*** Generic data storring in Neo4j finsihed for all plugins ***")

	return nil
}

// For testing purposes
/* // Print each InfrastructureComponent
for _, genericData := range genericData {
	logger.Debug(logger.LogFields(printInfrastructureComponent(&genericData)))
}func printInfrastructureComponent(ic *dataparser.InfrastructureComponent) {
	fmt.Println("InfrastructureComponent:")
	fmt.Println("----------------------------")
	fmt.Printf("ID: %s\n", ic.ID)
	fmt.Printf("Name: %s\n", ic.Name)
	fmt.Printf("Type: %s\n", ic.Type)
	fmt.Printf("AvailabilityZone: %s\n", ic.AvailabilityZone)
	fmt.Println("Metadata:")
	for key, value := range ic.Metadata {
		fmt.Printf("  %s: %v\n", key, value)
	}
	fmt.Println("Relationships:")
	for _, rel := range ic.Relationships {
		fmt.Printf("  Type: %s, Target: %s\n", rel.Type, rel.Target)
	}
	fmt.Println("----------------------------")
}
*/
// TODO make Manager trigger scan and traqnsformation concurrent
/* type TransformResult struct {
	Name   string
	Result YourResultType // Replace with the actual type of your results
}

type TransformError struct {
	Name string
	Err  error
}


// TODO add concurrency to scan and transform
func ConcurrentScanAndTransform() ([]TransformResult, []TransformError) {
	n := len(plugin.PluginRegistry)
	resultsCh := make(chan TransformResult, n)
	errorsCh := make(chan TransformError, n)

	for name, instance := range plugin.PluginRegistry {
		go func(pluginName string, pluginInstance YourPluginType) {
			// Scanning
			rawData := plugin.Scanner(pluginInstance)

			// Transformation
			transform := dataparser.TransformerFactory(pluginName)
			genericData, err := transform.Transform(rawData)
			if err != nil {
				errorsCh <- TransformError{Name: pluginName, Err: err}
				return
			}

			resultsCh <- TransformResult{Name: pluginName, Result: genericData}
		}(name, instance)
	}

	var results []TransformResult
	var errors []TransformError
	for i := 0; i < n; i++ {
		select {
		case result := <-resultsCh:
			results = append(results, result)
		case err := <-errorsCh:
			errors = append(errors, err)
		}
	}

	return results, errors
} */
