package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	v1 "k8s.io/api/core/v1"
)

func main() {
	fmt.Println("=== Kubernetes Client Usage Example ===")

	// kubernetes 클라이언트 생성
	fmt.Println("\n1. Creating Kubernetes client...")
	client := kubernetes.NewClient()
	ctx := context.Background()

	// 클라이언트 생성 후 간단한 테스트로 연결 확인
	_, err := client.GetNamespaces(ctx)
	if err != nil {
		log.Printf("Kubernetes client connection failed: %v", err)
		fmt.Println("Cannot connect to Kubernetes cluster. Please check your kubeconfig.")
		return
	}
	fmt.Println("✅ Kubernetes client created successfully.")

	// 2. Pod 목록 조회
	fmt.Println("\n2. Getting Pods...")
	pods, err := client.GetPods(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get pods: %v", err)
	} else {
		fmt.Printf("Found %d pods in default namespace\n", len(pods))
		for i, pod := range pods {
			if i < 3 { // 처음 3개만 출력
				fmt.Printf("  - Pod: %s (Status: %s)\n", pod.Name, pod.Status.Phase)
			}
		}
		if len(pods) > 3 {
			fmt.Printf("  ... and %d more pods\n", len(pods)-3)
		}
	}

	// 3. Service 목록 조회
	fmt.Println("\n3. Getting Services...")
	services, err := client.GetServices(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get services: %v", err)
	} else {
		fmt.Printf("Found %d services in default namespace\n", len(services))
		for i, service := range services {
			if i < 3 { // 처음 3개만 출력
				fmt.Printf("  - Service: %s (Type: %s)\n", service.Name, service.Spec.Type)
			}
		}
		if len(services) > 3 {
			fmt.Printf("  ... and %d more services\n", len(services)-3)
		}
	}

	// 4. StorageClass 목록 조회
	fmt.Println("\n4. Getting StorageClasses...")
	storageClasses, err := client.GetStorageClasses(ctx)
	if err != nil {
		log.Printf("Failed to get storage classes: %v", err)
	} else {
		fmt.Printf("Found %d storage classes\n", len(storageClasses))
		for i, sc := range storageClasses {
			if i < 3 { // 처음 3개만 출력
				fmt.Printf("  - StorageClass: %s (Provisioner: %s)\n", sc.Name, sc.Provisioner)
			}
		}
		if len(storageClasses) > 3 {
			fmt.Printf("  ... and %d more storage classes\n", len(storageClasses)-3)
		}
	}

	// 5. ConfigMap 목록 조회
	fmt.Println("\n5. Getting ConfigMaps...")
	configMaps, err := client.GetConfigMaps(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get config maps: %v", err)
	} else {
		fmt.Printf("Found %d config maps in default namespace\n", len(configMaps))
		for i, cm := range configMaps {
			if i < 3 { // 처음 3개만 출력
				fmt.Printf("  - ConfigMap: %s\n", cm.Name)
			}
		}
		if len(configMaps) > 3 {
			fmt.Printf("  ... and %d more config maps\n", len(configMaps)-3)
		}
	}

	// 6. Secret 목록 조회
	fmt.Println("\n6. Getting Secrets...")
	secrets, err := client.GetSecrets(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get secrets: %v", err)
	} else {
		fmt.Printf("Found %d secrets in default namespace\n", len(secrets))
		for i, secret := range secrets {
			if i < 3 { // 처음 3개만 출력
				fmt.Printf("  - Secret: %s (Type: %s)\n", secret.Name, secret.Type)
			}
		}
		if len(secrets) > 3 {
			fmt.Printf("  ... and %d more secrets\n", len(secrets)-3)
		}
	}

	// 7. Namespace 목록 조회
	fmt.Println("\n7. Getting Namespaces...")
	namespaces, err := client.GetNamespaces(ctx)
	if err != nil {
		log.Printf("Failed to get namespaces: %v", err)
	} else {
		fmt.Printf("Found %d namespaces\n", len(namespaces))
		for i, ns := range namespaces {
			if i < 5 { // 처음 5개만 출력
				fmt.Printf("  - Namespace: %s (Status: %s)\n", ns.Name, ns.Status.Phase)
			}
		}
		if len(namespaces) > 5 {
			fmt.Printf("  ... and %d more namespaces\n", len(namespaces)-5)
		}
	}

	// 8. 특정 Pod 조회 (첫 번째 Pod가 있는 경우)
	fmt.Println("\n8. Getting specific Pod details...")
	if len(pods) > 0 {
		podName := pods[0].Name
		pod, err := client.GetPod(ctx, "default", podName)
		if err != nil {
			log.Printf("Failed to get pod %s: %v", podName, err)
		} else {
			fmt.Printf("Pod details for %s:\n", podName)
			fmt.Printf("  - Name: %s\n", pod.Name)
			fmt.Printf("  - Namespace: %s\n", pod.Namespace)
			fmt.Printf("  - Status: %s\n", pod.Status.Phase)
			fmt.Printf("  - Node: %s\n", pod.Spec.NodeName)
			fmt.Printf("  - Creation Time: %s\n", pod.CreationTimestamp.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Println("No pods found to demonstrate specific pod retrieval")
	}

	// 9. 헬스 체크
	fmt.Println("\n9. Performing health check...")
	err = client.HealthCheck()
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("✅ Kubernetes client health check passed")
	}

	fmt.Println("\n=== Kubernetes Client Example Completed ===")
}

// demonstrateResourceFiltering : 리소스 필터링 데모
func demonstrateResourceFiltering(client kubernetes.Client, ctx context.Context) {
	fmt.Println("\n--- Resource Filtering Demo ---")

	// 라벨 셀렉터를 사용한 Pod 필터링
	fmt.Println("Filtering pods by label selector...")
	pods, err := client.GetPods(ctx, "default", "app=nginx")
	if err != nil {
		log.Printf("Failed to filter pods: %v", err)
	} else {
		fmt.Printf("Found %d pods with label app=nginx\n", len(pods))
	}

	// 특정 상태의 Pod 필터링
	fmt.Println("Filtering running pods...")
	runningPods, err := client.GetPods(ctx, "default", "")
	if err == nil {
		runningCount := 0
		for _, pod := range runningPods {
			if pod.Status.Phase == v1.PodRunning {
				runningCount++
			}
		}
		fmt.Printf("Found %d running pods\n", runningCount)
	}
}

// demonstrateResourceDetails : 리소스 상세 정보 데모
func demonstrateResourceDetails(client kubernetes.Client, ctx context.Context) {
	fmt.Println("\n--- Resource Details Demo ---")

	// Pod 상세 정보
	pods, err := client.GetPods(ctx, "default", "")
	if err == nil && len(pods) > 0 {
		pod := pods[0]
		fmt.Printf("Pod %s details:\n", pod.Name)
		fmt.Printf("  - UID: %s\n", pod.UID)
		fmt.Printf("  - Resource Version: %s\n", pod.ResourceVersion)
		fmt.Printf("  - Labels: %v\n", pod.Labels)
		fmt.Printf("  - Annotations: %v\n", pod.Annotations)

		// 컨테이너 정보
		fmt.Printf("  - Containers: %d\n", len(pod.Spec.Containers))
		for i, container := range pod.Spec.Containers {
			if i < 2 { // 처음 2개 컨테이너만 출력
				fmt.Printf("    - Container %d: %s (Image: %s)\n", i+1, container.Name, container.Image)
			}
		}
	}

	// Service 상세 정보
	services, err := client.GetServices(ctx, "default", "")
	if err == nil && len(services) > 0 {
		service := services[0]
		fmt.Printf("\nService %s details:\n", service.Name)
		fmt.Printf("  - Type: %s\n", service.Spec.Type)
		fmt.Printf("  - Cluster IP: %s\n", service.Spec.ClusterIP)
		fmt.Printf("  - Ports: %d\n", len(service.Spec.Ports))
		for i, port := range service.Spec.Ports {
			if i < 2 { // 처음 2개 포트만 출력
				fmt.Printf("    - Port %d: %d -> %d (%s)\n", i+1, port.Port, port.TargetPort.IntVal, port.Protocol)
			}
		}
	}
}
