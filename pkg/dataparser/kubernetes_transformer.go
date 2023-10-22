package dataparser

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	corev1 "k8s.io/api/core/v1"
)

// Transformer for Kubernetes
func (k *KubernetesTransformer) Transform(key string, data []interface{}) ([]InfrastructureComponent, error) {
	switch key {
	case "k8s_pv":
		k.pvcToPVMap = createPVCToPVMapFromRawData(data)
		return handlePV(data), nil
	case "k8s_node":
		return handleNode(data), nil
	case "k8s_pod":
		return handlePod(data, k.pvcToPVMap), nil
	default:
		return nil, fmt.Errorf("unknown key for OpenStack: %s", key)
	}

}

func handleNode(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, item := range data {
		node, ok := item.(corev1.Node)
		if !ok {
			fmt.Printf("Expected type v1.Node, but got: %T\n", item)
			continue
		}
		component := InfrastructureComponent{
			ID:   string(node.UID),
			Name: node.Name,
			Type: "ClusterNode",
			Metadata: map[string]interface{}{
				"CreatedAt": node.CreationTimestamp.Format(time.RFC3339),
			},
			Relationships: []Relationship{
				{
					Type:   "PROVISIONED_BY",
					Target: node.Status.NodeInfo.SystemUUID,
				},
			},
		}
		//logger.Debug(logger.LogFields{"node.Status.NodeInfo.SystemUUID": node.Status.NodeInfo.SystemUUID})
		components = append(components, component)
	}
	return components
}

func handlePod(data []interface{}, pvcToPVMap map[string]string) []InfrastructureComponent {
	var components []InfrastructureComponent
	seenPVCs := make(map[string]bool) // track the PVCs we've already created
	for _, item := range data {
		pod, ok := item.(corev1.Pod)
		if !ok {
			fmt.Printf("Expected type v1.Pod, but got: %T\n", item)
			continue
		}

		var volumeNames []string
		var podRelationships []Relationship

		// Relationship of the Pod to the Node
		podRelationships = append(podRelationships, Relationship{
			Type:   "RUNS_ON",
			Target: pod.Spec.NodeName,
		})

		// Process volumes and establish relationships to PVCs
		for _, volume := range pod.Spec.Volumes {
			volumeNames = append(volumeNames, volume.Name)

			if volume.PersistentVolumeClaim != nil {
				pvcName := volume.PersistentVolumeClaim.ClaimName
				pvName, exists := pvcToPVMap[pvcName]

				// Create PVC entity only if it hasn't been created before
				if exists && !seenPVCs[pvcName] {
					pvcComponent := InfrastructureComponent{
						ID:   pvcName,
						Name: pvcName,
						Type: "PersistentVolumeClaim",
						Relationships: []Relationship{
							{
								Type:   "BINDS_TO",
								Target: pvName,
							},
						},
					}
					//logger.Debug(logger.LogFields{"pvcComponent": pvcComponent})
					components = append(components, pvcComponent)
					seenPVCs[pvcName] = true
				}

				// Relationship of the Pod to the PVC
				podRelationships = append(podRelationships, Relationship{
					Type:   "USES_PVC",
					Target: pvcName,
				})
				//logger.Debug(logger.LogFields{"pvcName": pvcName})
			}
		}

		podComponent := InfrastructureComponent{
			ID:            string(pod.UID),
			Name:          pod.Name,
			Type:          "Pod",
			Metadata:      map[string]interface{}{"CreatedAt": pod.CreationTimestamp.Format(time.RFC3339), "Volumes": volumeNames},
			Relationships: podRelationships,
		}
		// If pod is attached with pd_annotations create transparency pod
		pdAnnotation, hasPD := pod.Annotations["has_pd"]
		// logger.Debug(logger.LogFields{"pdComponent ID": pdAnnotation})
		if hasPD {
			// Parse or validate JSON if necessary
			var pdData interface{}
			err := json.Unmarshal([]byte(pdAnnotation), &pdData)
			if err != nil {
				// Log or handle error
				logger.Error("Invalid JSON in pdAnnotation", err)
				continue // Skip this iteration
			}
			pdJSON, err := json.Marshal(pdData)
			if err != nil {
				// Log or handle error
				logger.Error("Error marshalling JSON for pdAnnotation", err)
				continue // Skip this iteration
			}

			pdComponent := InfrastructureComponent{
				ID:       "pd_indicator_" + string(pod.UID),
				Name:     "pd_indicator_" + pod.Name,
				Type:     "PDIndicator",
				Metadata: map[string]interface{}{"has_pd": string(pdJSON)},
			}

			components = append(components, pdComponent)
			// Create a relationship from pd_indicator to the Pod
			podComponent.Relationships = append(podComponent.Relationships, Relationship{
				Type:   "HAS_PD",
				Target: "pd_indicator_" + string(pod.UID),
			})
		}
		components = append(components, podComponent)
	}
	return components
}

func handlePV(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, item := range data {
		pv, ok := item.(corev1.PersistentVolume)
		if !ok {
			fmt.Printf("Expected type v1.PersistentVolume, but got: %T\n", item)
			continue
		}

		cinderVolumeID := pv.Spec.PersistentVolumeSource.Cinder.VolumeID

		component := InfrastructureComponent{
			ID:   string(pv.UID),
			Name: pv.Name,
			Type: "PersistentVolume",
			Metadata: map[string]interface{}{
				"CreatedAt": pv.CreationTimestamp.Format(time.RFC3339),
				// Add any other relevant metadata here
			},
			Relationships: []Relationship{
				{
					Type:   "STORED_ON",
					Target: cinderVolumeID,
				},
			},
		}
		//logger.Debug(logger.LogFields{"PV": pv})
		components = append(components, component)
	}
	return components
}

func createPVCToPVMapFromRawData(data []interface{}) map[string]string {
	pvcToPV := make(map[string]string)
	for _, item := range data {
		if pv, ok := item.(corev1.PersistentVolume); ok && pv.Spec.ClaimRef != nil {
			pvcToPV[pv.Spec.ClaimRef.Name] = pv.Name

		}

	}
	return pvcToPV
}
