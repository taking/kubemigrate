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
	client, err := kubernetes.NewClient()
	if err != nil {
		log.Printf("Failed to create Kubernetes client: %v", err)
		fmt.Println("Cannot create Kubernetes client. Please check your kubeconfig.")
		return
	}
	ctx := context.Background()

	// 클라이언트 생성 후 간단한 테스트로 연결 확인
	_, err = client.GetNamespaces(ctx, "")
	if err != nil {
		log.Printf("Kubernetes client connection failed: %v", err)
		fmt.Println("Cannot connect to Kubernetes cluster. Please check your kubeconfig.")
		return
	}
	fmt.Println("✅ Kubernetes client created successfully.")

	// 2. Pod 목록 조회
	fmt.Println("\n2. Getting Pods...")
	podsResponse, err := client.GetPods(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get pods: %v", err)
	} else {
		// Type assertion to get the actual PodList
		podList, ok := podsResponse.(*v1.PodList)
		if !ok {
			log.Printf("Failed to cast response to PodList")
		} else {
			pods := podList.Items
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
	}

	// 3. ConfigMap 목록 조회
	fmt.Println("\n3. Getting ConfigMaps...")
	configMapsResponse, err := client.GetConfigMaps(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get config maps: %v", err)
	} else {
		// Type assertion to get the actual ConfigMapList
		configMapList, ok := configMapsResponse.(*v1.ConfigMapList)
		if !ok {
			log.Printf("Failed to cast response to ConfigMapList")
		} else {
			configMaps := configMapList.Items
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
	}

	// 4. StorageClass 목록 조회
	fmt.Println("\n4. Getting StorageClasses...")
	storageClassesResponse, err := client.GetStorageClasses(ctx, "")
	if err != nil {
		log.Printf("Failed to get storage classes: %v", err)
	} else {
		// Type assertion to get the actual StorageClassList
		storageClassList, ok := storageClassesResponse.(*storagev1.StorageClassList)
		if !ok {
			log.Printf("Failed to cast response to StorageClassList")
		} else {
			storageClasses := storageClassList.Items
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
	}

	// 5. Secret 목록 조회
	fmt.Println("\n5. Getting Secrets...")
	secretsResponse, err := client.GetSecrets(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get secrets: %v", err)
	} else {
		// Type assertion to get the actual SecretList
		secretList, ok := secretsResponse.(*v1.SecretList)
		if !ok {
			log.Printf("Failed to cast response to SecretList")
		} else {
			secrets := secretList.Items
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
	}

	// 6. Namespace 목록 조회
	fmt.Println("\n6. Getting Namespaces...")
	namespacesResponse, err := client.GetNamespaces(ctx, "")
	if err != nil {
		log.Printf("Failed to get namespaces: %v", err)
	} else {
		// Type assertion to get the actual NamespaceList
		namespaceList, ok := namespacesResponse.(*v1.NamespaceList)
		if !ok {
			log.Printf("Failed to cast response to NamespaceList")
		} else {
			namespaces := namespaceList.Items
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
	}

	// 8. 특정 Pod 조회 (첫 번째 Pod가 있는 경우)
	fmt.Println("\n8. Getting specific Pod details...")
	// Get pods again for specific pod retrieval
	podsResponse, err = client.GetPods(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to get pods for specific retrieval: %v", err)
	} else {
		podList, ok := podsResponse.(*v1.PodList)
		if !ok {
			log.Printf("Failed to cast response to PodList for specific retrieval")
		} else {
			pods := podList.Items
			if len(pods) > 0 {
				podName := pods[0].Name
				podResponse, err := client.GetPods(ctx, "default", podName)
				if err != nil {
					log.Printf("Failed to get pod %s: %v", podName, err)
				} else {
					// Type assertion for single pod
					pod, ok := podResponse.(*v1.Pod)
					if !ok {
						log.Printf("Failed to cast response to Pod")
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
	}

	// 7. 클라이언트 연결 상태 확인
	fmt.Println("\n7. Verifying client connection...")
	_, err = client.GetNamespaces(ctx, "")
	if err != nil {
		log.Printf("Connection verification failed: %v", err)
	} else {
		fmt.Println("✅ Kubernetes client connection verified")
	}

	fmt.Println("\n=== Kubernetes Client Example Completed ===")
}

// demonstrateResourceFiltering : 리소스 필터링 데모
func demonstrateResourceFiltering(client kubernetes.Client, ctx context.Context) {
	fmt.Println("\n--- Resource Filtering Demo ---")

	// 라벨 셀렉터를 사용한 Pod 필터링
	fmt.Println("Filtering pods by label selector...")
	podsResponse, err := client.GetPods(ctx, "default", "")
	if err != nil {
		log.Printf("Failed to filter pods: %v", err)
	} else {
		podList, ok := podsResponse.(*v1.PodList)
		if !ok {
			log.Printf("Failed to cast response to PodList")
		} else {
			pods := podList.Items
			fmt.Printf("Found %d pods with label app=nginx\n", len(pods))
		}
	}

	// 특정 상태의 Pod 필터링
	fmt.Println("Filtering running pods...")
	runningPodsResponse, err := client.GetPods(ctx, "default", "")
	if err == nil {
		podList, ok := runningPodsResponse.(*v1.PodList)
		if ok {
			runningPods := podList.Items
			runningCount := 0
			for _, pod := range runningPods {
				if pod.Status.Phase == v1.PodRunning {
					runningCount++
				}
			}
			fmt.Printf("Found %d running pods\n", runningCount)
		}
	}
}

// demonstrateResourceDetails : 리소스 상세 정보 데모
func demonstrateResourceDetails(client kubernetes.Client, ctx context.Context) {
	fmt.Println("\n--- Resource Details Demo ---")

	// Pod 상세 정보
	podsResponse, err := client.GetPods(ctx, "default", "")
	if err == nil {
		podList, ok := podsResponse.(*v1.PodList)
		if ok && len(podList.Items) > 0 {
			pod := podList.Items[0]
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
	}

	// ConfigMap 상세 정보
	configMapsResponse, err := client.GetConfigMaps(ctx, "default", "")
	if err == nil {
		configMapList, ok := configMapsResponse.(*v1.ConfigMapList)
		if ok && len(configMapList.Items) > 0 {
			configMap := configMapList.Items[0]
			fmt.Printf("\nConfigMap %s details:\n", configMap.Name)
			fmt.Printf("  - Data keys: %d\n", len(configMap.Data))
			fmt.Printf("  - Binary data keys: %d\n", len(configMap.BinaryData))
			for key := range configMap.Data {
				fmt.Printf("    - Data key: %s\n", key)
				break // 첫 번째 키만 출력
			}
		}
	}
}
