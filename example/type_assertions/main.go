package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/utils"
)

func main() {
	fmt.Println("=== Type Assertion Helper Functions Example ===")
	fmt.Println("This example demonstrates how to use helper functions for type assertions.")

	ctx := context.Background()

	// Initialize clients
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		log.Printf("Failed to create Kubernetes client: %v", err)
		return
	}

	minioClient, err := minio.NewClient()
	if err != nil {
		log.Printf("Failed to create MinIO client: %v", err)
		return
	}

	// Example 1: Using helper functions for Pods
	fmt.Println("\n1. Getting Pods with helper functions...")
	podsResponse, err := k8sClient.GetPods(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get pods: %v", err)
	} else {
		// Use helper function instead of manual type assertion
		pods, err := utils.ExtractPods(podsResponse)
		if err != nil {
			log.Printf("Failed to extract pods: %v", err)
		} else {
			fmt.Printf("Found %d pods in default namespace\n", len(pods))
			for i, pod := range pods {
				if i < 3 { // Show first 3 pods
					fmt.Printf("  - Pod: %s (Status: %s)\n", pod.Name, pod.Status.Phase)
				}
			}
			if len(pods) > 3 {
				fmt.Printf("  ... and %d more pods\n", len(pods)-3)
			}
		}
	}

	// Example 2: Using helper functions for ConfigMaps
	fmt.Println("\n2. Getting ConfigMaps with helper functions...")
	configMapsResponse, err := k8sClient.GetConfigMaps(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get config maps: %v", err)
	} else {
		configMaps, err := utils.ExtractConfigMaps(configMapsResponse)
		if err != nil {
			log.Printf("Failed to extract config maps: %v", err)
		} else {
			fmt.Printf("Found %d config maps in default namespace\n", len(configMaps))
			for i, configMap := range configMaps {
				if i < 3 { // Show first 3 config maps
					fmt.Printf("  - ConfigMap: %s\n", configMap.Name)
				}
			}
			if len(configMaps) > 3 {
				fmt.Printf("  ... and %d more config maps\n", len(configMaps)-3)
			}
		}
	}

	// Example 3: Using helper functions for MinIO buckets
	fmt.Println("\n3. Getting MinIO buckets with helper functions...")
	bucketsResponse, err := minioClient.ListBuckets(ctx)
	if err != nil {
		log.Printf("Failed to list buckets: %v", err)
	} else {
		buckets, err := utils.ExtractBuckets(bucketsResponse)
		if err != nil {
			log.Printf("Failed to extract buckets: %v", err)
		} else {
			fmt.Printf("Found %d buckets in MinIO\n", len(buckets))
			for i, bucket := range buckets {
				if i < 3 { // Show first 3 buckets
					fmt.Printf("  - Bucket: %s (Created: %s)\n", bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
				}
			}
			if len(buckets) > 3 {
				fmt.Printf("  ... and %d more buckets\n", len(buckets)-3)
			}
		}
	}

	// Example 4: Using helper functions for single resource retrieval
	fmt.Println("\n4. Getting a specific Pod with helper functions...")
	// First get a pod name
	podsResponse, err = k8sClient.GetPods(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get pods for specific retrieval: %v", err)
	} else {
		pods, err := utils.ExtractPods(podsResponse)
		if err != nil {
			log.Printf("Failed to extract pods: %v", err)
		} else if len(pods) > 0 {
			podName := pods[0].Name
			podResponse, err := k8sClient.GetPods(ctx, "default", podName)
			if err != nil {
				log.Printf("Failed to get pod %s: %v", podName, err)
			} else {
				pod, err := utils.ExtractPod(podResponse)
				if err != nil {
					log.Printf("Failed to extract pod: %v", err)
				} else {
					fmt.Printf("Pod details for %s:\n", podName)
					fmt.Printf("  - Name: %s\n", pod.Name)
					fmt.Printf("  - Namespace: %s\n", pod.Namespace)
					fmt.Printf("  - Status: %s\n", pod.Status.Phase)
					fmt.Printf("  - Node: %s\n", pod.Spec.NodeName)
					fmt.Printf("  - Creation Time: %s\n", pod.CreationTimestamp.Format("2006-01-02 15:04:05"))
				}
			}
		} else {
			fmt.Println("No pods found to demonstrate specific pod retrieval")
		}
	}

	fmt.Println("\n=== Type Assertion Helper Functions Example Completed ===")
}
