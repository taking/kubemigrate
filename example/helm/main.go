package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taking/kubemigrate/pkg/client/helm"
)

func main() {
	fmt.Println("=== Helm Client Usage Example ===")

	// 1. Helm 클라이언트 생성
	fmt.Println("\n1. Creating Helm client...")
	client := helm.NewClient()
	ctx := context.Background()

	if err := client.HealthCheck(ctx); err != nil {
		log.Printf("Helm client health check failed: %v", err)
		fmt.Println("Cannot connect to Kubernetes cluster. Please check your kubeconfig.")
		return
	}
	fmt.Println("✅ Helm client created successfully.")

	// 2. 차트 목록 조회
	fmt.Println("\n2. Retrieving installed Helm charts...")
	namespace := "default"
	charts, err := client.GetCharts(ctx, namespace)
	if err != nil {
		log.Printf("Failed to retrieve chart list: %v", err)
	} else {
		fmt.Printf("✅ Found %d charts in '%s' namespace:\n", len(charts), namespace)
		for i, chart := range charts {
			if i < 5 { // 처음 5개만 출력
				fmt.Printf("  - %s (version: %s, status: %s)\n",
					chart.Name,
					chart.Chart.Metadata.Version,
					chart.Info.Status)
			}
		}
		if len(charts) > 5 {
			fmt.Printf("  ... and %d more\n", len(charts)-5)
		}
	}

	// 3. 특정 차트 조회 (첫 번째 차트가 있다면)
	if len(charts) > 0 {
		firstChart := charts[0]
		fmt.Printf("\n3. Retrieving detailed information for chart '%s'...\n", firstChart.Name)

		chart, err := client.GetChart(ctx, firstChart.Name, namespace, firstChart.Version)
		if err != nil {
			log.Printf("Failed to retrieve chart: %v", err)
		} else {
			fmt.Printf("✅ Chart details:\n")
			fmt.Printf("  - Name: %s\n", chart.Name)
			fmt.Printf("  - Chart Name: %s\n", chart.Chart.Metadata.Name)
			fmt.Printf("  - Version: %s\n", chart.Chart.Metadata.Version)
			fmt.Printf("  - App Version: %s\n", chart.Chart.Metadata.AppVersion)
			fmt.Printf("  - Status: %s\n", chart.Info.Status)
			fmt.Printf("  - Namespace: %s\n", chart.Namespace)
		}
	}

	// 4. 차트 설치 여부 확인
	testReleaseName := "test-release"
	fmt.Printf("\n4. Checking if chart '%s' is installed...\n", testReleaseName)

	installed, release, err := client.IsChartInstalled(testReleaseName)
	if err != nil {
		log.Printf("Failed to check chart installation status: %v", err)
	} else {
		if installed {
			fmt.Printf("✅ Chart '%s' is installed.\n", testReleaseName)
			if release != nil {
				fmt.Printf("  - Version: %s\n", release.Chart.Metadata.Version)
				fmt.Printf("  - Status: %s\n", release.Info.Status)
			}
		} else {
			fmt.Printf("ℹ️  Chart '%s' is not installed.\n", testReleaseName)
		}
	}

	// 5. 여러 네임스페이스에서 차트 조회
	fmt.Println("\n5. Retrieving charts from multiple namespaces...")

	namespaces := []string{"default", "kube-system", "kube-public"}
	for _, ns := range namespaces {
		charts, err := client.GetCharts(ctx, ns)
		if err != nil {
			log.Printf("Failed to retrieve charts from namespace '%s': %v", ns, err)
		} else {
			fmt.Printf("✅ Namespace '%s': %d charts\n", ns, len(charts))
		}
	}

	// 6. 헬스 체크
	fmt.Println("\n6. Final health check...")
	if err := client.HealthCheck(ctx); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("✅ Helm client is working properly.")
	}

	fmt.Println("\n=== Helm Client Example Completed ===")
}
