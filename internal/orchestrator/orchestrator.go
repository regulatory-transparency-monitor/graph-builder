package orchestrator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/plugin"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/repository"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/robfig/cron"
)

type Orchestrator struct {
	Transformers   map[string]dataparser.Transformer
	Neo4jRepo      repository.Repository
	CurrentVersion string
}

func NewOrchestrator(tf map[string]dataparser.Transformer, repo repository.Repository) *Orchestrator {
	o := &Orchestrator{
		Transformers:   tf,
		Neo4jRepo:      repo,
		CurrentVersion: "0.0.0", // default
	}
	o.initVersioning() // Set the version based on database or initialize it
	return o
}

func (o *Orchestrator) initVersioning() {
	// Try to fetch the latest version from the database

	// If there's an error, it might be because the Metadata node doesn't exist
	version, err := o.Neo4jRepo.GetLatestVersion()
	if err != nil {
		logger.Warning("Couldn't fetch latest version, initializing with version 0.0.1")

		o.CurrentVersion = "0.0.1"
	} else {
		o.CurrentVersion = version
	}

	logger.Info("Initialized with version: ", logger.LogFields{"version": o.CurrentVersion})
}

func getCurrentTimeString() string {
	// Assuming you want a specific time format, like "2006-01-02 15:04:05"
	return time.Now().Format("2006-01-02 15:04:05")
}

func (o *Orchestrator) incrementVersion() {
	o.CurrentVersion = incrementVersion(o.CurrentVersion)
}

func incrementVersion(version string) string {
	splitVersion := strings.Split(version, ".")
	if len(splitVersion) != 3 {
		logger.Error("Invalid version format: %s", version)
		return "0.0.0" // default if version format is not as expected
	}

	major, err1 := strconv.Atoi(splitVersion[0])
	minor, err2 := strconv.Atoi(splitVersion[1])
	patch, err3 := strconv.Atoi(splitVersion[2])

	if err1 != nil || err2 != nil || err3 != nil {
		logger.Error("Failed to parse version components: ", logger.LogFields{"major": err1, "minor": err2, "patch": err3})
		return "0.0.0"
	}
	// Incrementing only the patch version
	patch++

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func (o *Orchestrator) Run() error {

	// 1) Create the initial version node
	err := o.Neo4jRepo.SetupUUIDForKnownLabels()
	if err != nil {
		logger.Error("Failed to create UUID constraints: %v", err)
		return err
	}

	// 2) Run Initial infrastructure scan
	err = o.getInfrastructure()
	if err != nil {
		return err
	}

	// Only used wehne nodes are created
	/* labels, err := o.Neo4jRepo.GetLabels()
	if err != nil {
		logger.Error("Failed to get labels: %v", err)
		return err
	} */

	logger.Info("*** Orchestrator started successfully ***")

	// 3) Start periodic scans
	o.startPeriodicScans()

	return nil
}

// PeriodicScan scans the infrastructure periodically using provider plugins
func (o *Orchestrator) startPeriodicScans() {
	c := cron.New()
	c.AddFunc("@every 30s", func() {
		o.incrementVersion()
		o.getInfrastructure()
	})
	c.Start()
}

func (o *Orchestrator) getInfrastructure() error {
	err := o.Neo4jRepo.CreateMetadataNode(o.CurrentVersion, getCurrentTimeString())
	if err != nil {
		logger.Error("Failed to create metadata node: %v", err)
		return err
	}

	// 1) Start scanning resources for each provider plugin enabled
	for providerType, instance := range plugin.PluginRegistry {
		logger.Info("Fetching API services using ", logger.LogFields{"provider plugin": providerType})
		rawDataMap := plugin.Scanner(instance) // returns map[string][]interface{}
		logger.Info("API services fetched successfully")
		// 2) Transform raw data into generic data using the appropriate transformer
		genericData, err := dataparser.TransformData(rawDataMap) // Call the TransformData function here
		if err != nil {
			logger.Error("Error transforming data: %v", err)
			continue // continue to next provider if there's an error
		}
		logger.Info("Generic data transformed")
		// 3) Store generic data in Neo4j
		var projectUUID string
		for _, component := range genericData {
			switch component.Type {
			case "Project":
				projectUUID, err = o.Neo4jRepo.CreateOrUpdateProject(component)
				if err != nil {
					logger.Error("Error storing project in Neo4j: %v", err)
				}
				logger.Debug("Project UUID after creating node in orchestrator: ", logger.LogFields{"uuid": projectUUID})
			case "Server":
				err := o.Neo4jRepo.CreateOrUpdateServer(component)
				if err != nil {
					logger.Error("Error storing server in Neo4j: %v", err)
				}
			case "Volume":
				err := o.Neo4jRepo.CreateOrUpdateVolume(component)
				if err != nil {
					logger.Error("Error storing volume in Neo4j: %v", err)
				}
			case "ClusterNode":
				err := o.Neo4jRepo.CreateOrUpdateClusterNode(component)
				if err != nil {
					logger.Error("Error storing clusterNode in Neo4j: %v", err)
				}
			case "Pod":
				err := o.Neo4jRepo.CreateOrUpdatePod(component)
				if err != nil {
					logger.Error("Error storing pod in Neo4j: %v", err)
				}

			}

		}
		err = o.Neo4jRepo.LinkResourceToMetadata(o.CurrentVersion, projectUUID)
		if err != nil {
			logger.Error("Failed to link resource to metadata: %v", err)
		}
		logger.Info("*** Generic data stored in Neo4j ***")

	}

	return nil
}

/* // Print each InfrastructureComponent
for _, genericData := range genericData {
	logger.Debug(logger.LogFields(printInfrastructureComponent(&genericData)))
} */

func printInfrastructureComponent(ic *dataparser.InfrastructureComponent) {
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

// TODO make orchestrator trigger scan and traqnsformation concurrent
/* type TransformResult struct {
	Name   string
	Result YourResultType // Replace with the actual type of your results
}

type TransformError struct {
	Name string
	Err  error
}

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

/* // Fetch the required plugin instance
openstackPlugin, err := plugin.GetPluginInstance("openstack")
if err != nil {
	logger.Fatal("Failed to retrieve plugin instance: %v", err)
}

// Call the Scanner function
data := plugin.Scanner(openstackPlugin) */
