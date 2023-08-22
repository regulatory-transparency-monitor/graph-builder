package orchestrator

import (
	"fmt"

	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/plugin"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/repository"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

type Orchestrator struct {
	TransformerFactory dataparser.TransformerFactory
	Neo4jRepo          repository.Repository
}

func NewOrchestrator(tf dataparser.TransformerFactory, repo repository.Repository) *Orchestrator {
	return &Orchestrator{
		TransformerFactory: tf,
		Neo4jRepo:          repo,
	}
}

func (o *Orchestrator) Run() error {
	// 1) Initialize plugin constructor functions
	plugin.InitConstructor()

	// 2) Initialize and Register enabled plugins, e.g. openstack, kubernetes
	plugin.RegisterPlugin()

	// 3) Scan provider resources to fetch data
	for name, instance := range plugin.PluginRegistry {
		logger.Info("Scanning resources", logger.LogFields{"provider plugin ": name})
		rawData := plugin.Scanner(instance)

		// 4) Transform raw data into generic data
		transformer := o.TransformerFactory.GetTransformer(name)

		genericData, err := transformer.Transform(rawData)
		if err != nil {
			logger.Error("Error transforming data: %v", err)
		}

		// Print each InfrastructureComponent
		for _, genericData := range genericData {
			printInfrastructureComponent(&genericData)
		}

		// TODO send generic Data to repository and store nodes and relationships in neo4j
		// 4) Store generic data in Neo4j
		for _, component := range genericData {
		    if component.Type == "Project" {
		        err := a.Neo4jRepo.CreateOrUpdateProject(component)
		        if err != nil {
		            logger.Error("Error storing project in Neo4j: %v", err)
		        }
		    }
		    // TODO: Handle other component types as needed
		}

	}
	return nil

}

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
