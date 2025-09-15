package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
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
	fmt.Println("\n2. Retrieving Pod list...")
	namespace := "default"
	pods, err := client.GetPods(ctx, namespace, "")
	if err != nil {
		log.Printf("Failed to retrieve Pod list: %v", err)
	} else {
		if podList, ok := pods.(*v1.PodList); ok {
			fmt.Printf("✅ Found %d Pods in '%s' namespace:\n", len(podList.Items), namespace)
			for i, pod := range podList.Items {
				if i < 5 { // 처음 5개만 출력
					fmt.Printf("  - %s (status: %s, node: %s)\n",
						pod.Name,
						string(pod.Status.Phase),
						pod.Spec.NodeName)
				}
			}
			if len(podList.Items) > 5 {
				fmt.Printf("  ... and %d more\n", len(podList.Items)-5)
			}
		}
	}

	// 3. ConfigMap 목록 조회
	fmt.Println("\n3. Retrieving ConfigMap list...")
	configMaps, err := client.GetConfigMaps(ctx, namespace, "")
	if err != nil {
		log.Printf("Failed to retrieve ConfigMap list: %v", err)
	} else {
		if cmList, ok := configMaps.(*v1.ConfigMapList); ok {
			fmt.Printf("✅ Found %d ConfigMaps in '%s' namespace:\n", len(cmList.Items), namespace)
			for i, cm := range cmList.Items {
				if i < 5 { // 처음 5개만 출력
					fmt.Printf("  - %s\n", cm.Name)
				}
			}
			if len(cmList.Items) > 5 {
				fmt.Printf("  ... and %d more\n", len(cmList.Items)-5)
			}
		}
	}

	// 4. StorageClass 목록 조회
	fmt.Println("\n4. Retrieving StorageClass list...")
	storageClasses, err := client.GetStorageClasses(ctx, "")
	if err != nil {
		log.Printf("Failed to retrieve StorageClass list: %v", err)
	} else {
		if scList, ok := storageClasses.(*storagev1.StorageClassList); ok {
			fmt.Printf("✅ Found %d StorageClasses in cluster:\n", len(scList.Items))
			for i, sc := range scList.Items {
				if i < 5 { // 처음 5개만 출력
					fmt.Printf("  - %s (provisioner: %s)\n",
						sc.Name,
						sc.Provisioner)
				}
			}
			if len(scList.Items) > 5 {
				fmt.Printf("  ... and %d more\n", len(scList.Items)-5)
			}
		}
	}

	// 5. 네임스페이스 목록 조회
	fmt.Println("\n5. Retrieving namespace list...")
	namespaces, err := client.GetNamespaces(ctx)
	if err != nil {
		log.Printf("Failed to retrieve namespace list: %v", err)
	} else {
		fmt.Printf("✅ Found %d namespaces in cluster:\n", len(namespaces.Items))
		for i, ns := range namespaces.Items {
			if i < 5 { // 처음 5개만 출력
				fmt.Printf("  - %s\n", ns.Name)
			}
		}
		if len(namespaces.Items) > 5 {
			fmt.Printf("  ... and %d more\n", len(namespaces.Items)-5)
		}
	}

	fmt.Println("\n=== Kubernetes Client Example Completed ===")
}
