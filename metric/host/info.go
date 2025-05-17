// Package host provides utilities for collecting host and container information
package host

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/MichaelAJay/go-metrics/metric"
)

// Info represents host and container information
type Info struct {
	Hostname      string
	OS            string
	Architecture  string
	CPUCores      int
	ContainerID   string
	KubeNode      string
	KubePod       string
	KubeNamespace string
	Region        string
	Zone          string
	Environment   string
}

// NewInfo gathers host information
func NewInfo() (*Info, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	info := &Info{
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUCores:     runtime.NumCPU(),
		Environment:  getEnv("ENVIRONMENT", "development"),
		Region:       getEnv("REGION", ""),
		Zone:         getEnv("ZONE", ""),
	}

	// Try to detect container environment
	info.detectContainer()
	info.detectKubernetes()

	return info, nil
}

// AsMetricTags converts host info to metric tags
func (i *Info) AsMetricTags() metric.Tags {
	tags := metric.Tags{
		"host":        i.Hostname,
		"os":          i.OS,
		"arch":        i.Architecture,
		"environment": i.Environment,
	}

	// Only add non-empty values
	if i.ContainerID != "" {
		tags["container_id"] = i.ContainerID
	}
	if i.KubeNode != "" {
		tags["kube_node"] = i.KubeNode
	}
	if i.KubePod != "" {
		tags["kube_pod"] = i.KubePod
	}
	if i.KubeNamespace != "" {
		tags["kube_namespace"] = i.KubeNamespace
	}
	if i.Region != "" {
		tags["region"] = i.Region
	}
	if i.Zone != "" {
		tags["zone"] = i.Zone
	}

	return tags
}

// detectContainer attempts to detect if running in a container
func (i *Info) detectContainer() {
	// Simple detection method - check if cgroup file exists and contains docker/containerd/etc.
	cgroupData, err := os.ReadFile("/proc/self/cgroup")
	if err == nil {
		content := string(cgroupData)

		// Docker containers often have a pattern like "docker-<container_id>.scope"
		if idx := strings.Index(content, "docker-"); idx != -1 {
			idEnd := strings.Index(content[idx:], ".scope")
			if idEnd > 0 {
				i.ContainerID = content[idx+7 : idx+idEnd]
			}
		}

		// If containerID is still empty, look for any ID-like patterns
		if i.ContainerID == "" && (strings.Contains(content, "docker") ||
			strings.Contains(content, "containerd") ||
			strings.Contains(content, "cri-o")) {
			// This is a simplified detection - a real implementation would need more robust parsing
			i.ContainerID = "detected-but-unknown-id"
		}
	}

	// Alternative: check for .dockerenv file which exists in Docker containers
	if _, err := os.Stat("/.dockerenv"); err == nil {
		if i.ContainerID == "" {
			i.ContainerID = "docker-container"
		}
	}
}

// detectKubernetes attempts to detect if running in Kubernetes
func (i *Info) detectKubernetes() {
	// In Kubernetes, these are usually set as environment variables
	i.KubeNode = getEnv("NODE_NAME", "")
	i.KubePod = getEnv("POD_NAME", "")
	i.KubeNamespace = getEnv("POD_NAMESPACE", "")

	// If not explicitly set, try to get pod name from hostname
	// Kubernetes sets the hostname to the pod name by default
	if i.KubePod == "" && i.ContainerID != "" {
		// K8s pod hostnames follow a pattern that can often be detected
		hostname, _ := os.Hostname()
		if strings.Count(hostname, "-") >= 2 {
			i.KubePod = hostname
		}
	}
}

// InjectHostInfo adds host information tags to a Registry
func InjectHostInfo(registry metric.Registry) error {
	info, err := NewInfo()
	if err != nil {
		return fmt.Errorf("failed to collect host info: %w", err)
	}

	// Create a gauge that indicates the service is running and attach host info as tags
	gauge := registry.Gauge(metric.Options{
		Name:        "service_info",
		Description: "Information about the running service instance",
		Unit:        "",
		Tags:        info.AsMetricTags(),
	})

	// Set to 1 to indicate the service is up
	gauge.Set(1)

	return nil
}

// Helper functions

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
