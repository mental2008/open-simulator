package utils

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	api "k8s.io/kubernetes/pkg/apis/core"

	simontype "github.com/alibaba/open-simulator/pkg/type"
	"github.com/alibaba/open-simulator/pkg/type/open-gpu-share/utils"
)

const (
	ResourceMilliCpu = "MilliCpu"
	ResourceMemory   = "Memory"
	ResourceGpu      = "Gpu"
	ResourceMilliGpu = "MilliGpu"
)

var resourceList = []string{ResourceMilliCpu, ResourceMemory, ResourceGpu, ResourceMilliGpu}

type ResourceSummary struct {
	name        string
	requested   int64
	allocatable int64
}

type AllocAmount struct {
	NodeName    string
	Requested   map[string]int64
	Allocatable map[string]int64
}

func (a AllocAmount) Add(b AllocAmount) error {
	if len(a.Requested) == 0 && len(a.Allocatable) == 0 {
		for k, v := range b.Requested {
			a.Requested[k] = v
		}
		for k, v := range b.Allocatable {
			a.Allocatable[k] = v
		}
		return nil
	}
	if len(b.Requested) == 0 && len(b.Allocatable) == 0 {
		return nil
	}
	if len(a.Requested) != len(b.Requested) || len(a.Allocatable) != len(b.Allocatable) {
		return fmt.Errorf("len(a)(%d, %d) != len(b)(%d, %d)",
			len(a.Requested), len(b.Requested), len(a.Allocatable), len(b.Allocatable))
	}

	for k, _ := range a.Requested {
		_, ok := a.Allocatable[k]
		if !ok {
			return fmt.Errorf("%s not in a.Allocatable", k)
		}
		ba, ok := b.Allocatable[k]
		if !ok {
			return fmt.Errorf("%s not in b.Allocatable", k)
		}
		br, ok := b.Requested[k]
		if !ok {
			return fmt.Errorf("%s not in b.Requested", k)
		}

		a.Requested[k] += br
		a.Allocatable[k] += ba
	}
	return nil
}

func ReportNodeAllocationRate(aamap map[string]AllocAmount, verbose int) (rs []ResourceSummary) {
	requested := make(map[string]int64)
	allocatable := make(map[string]int64)
	clusterAllocAmount := AllocAmount{"cluster", requested, allocatable}

	for _, amount := range aamap {
		clusterAllocAmount.Add(amount)
	}

	if verbose >= 1 {
		fmt.Printf("Allocation Ratio:\n")
	}
	for _, k := range resourceList {
		rval := clusterAllocAmount.Requested[k]
		aval := clusterAllocAmount.Allocatable[k]
		ratio := 100.0 * float64(rval) / float64(aval)
		if verbose >= 1 {
			fmt.Printf("    %-8s: %4.1f%% (%d/%d)\n", k, ratio, rval, aval)
		}
		rs = append(rs, ResourceSummary{k, rval, aval})
	}
	return rs
}

func GetNodeAllocMap(nodeStatus []simontype.NodeStatus) (map[string]AllocAmount, error) {
	allPods := GetAllPodsPtrFromNodeStatus(nodeStatus)

	nodeAllocMap := make(map[string]AllocAmount)
	for _, ns := range nodeStatus {
		node := ns.Node
		allocatable := make(map[string]int64)
		for k, v := range node.Status.Allocatable {
			if k.String() == api.ResourceCPU.String() {
				allocatable[ResourceMilliCpu] = v.MilliValue()
			} else if k.String() == api.ResourceMemory.String() {
				allocatable[ResourceMemory] = v.Value()
			}
		}
		gpuNumber := utils.GetGpuCountOfNode(node)
		allocatable[ResourceGpu] = int64(gpuNumber)
		allocatable[ResourceMilliGpu] = int64(gpuNumber) * utils.MILLI

		requested := make(map[string]int64)
		reqs, _ := GetPodsTotalRequestsAndLimitsByNodeName(allPods, node.Name)
		nodeCpuReq, nodeMemReq := reqs[corev1.ResourceCPU], reqs[corev1.ResourceMemory]
		requested[ResourceMilliCpu] = nodeCpuReq.MilliValue()
		requested[ResourceMemory] = nodeMemReq.Value()

		requested[ResourceMilliGpu] = 0
		requested[ResourceGpu] = 0
		if gpuNodeInfoStr, err := GetGpuNodeInfoFromAnnotation(node); err == nil && gpuNodeInfoStr != nil {
			for _, dev := range gpuNodeInfoStr.DevsBrief {
				if dev.GpuUsedMilli > 0 {
					requested[ResourceGpu] += 1
				}
				requested[ResourceMilliGpu] += dev.GpuUsedMilli
			}
		}
		nodeAllocMap[node.Name] = AllocAmount{node.Name, requested, allocatable}
	}
	return nodeAllocMap, nil
}
