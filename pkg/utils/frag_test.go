package utils

import (
	simontype "github.com/alibaba/open-simulator/pkg/type"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestingGenerateGetTypicalPods() simontype.TargetPodList {
	typicalPods := simontype.TargetPodList{}
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 6000, MilliGpu: 465, GpuNumber: 1, GpuType: ""}, Percentage: 9.33 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 8000, MilliGpu: 440, GpuNumber: 1, GpuType: "2080"}, Percentage: 9.15 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 8000, MilliGpu: 475, GpuNumber: 1, GpuType: "T4"}, Percentage: 8.76 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 8000, MilliGpu: 440, GpuNumber: 1, GpuType: "P100"}, Percentage: 8.72 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 2000, MilliGpu: 465, GpuNumber: 1, GpuType: ""}, Percentage: 8.68 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 12000, MilliGpu: 900, GpuNumber: 1, GpuType: ""}, Percentage: 8.65 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 4000, MilliGpu: 900, GpuNumber: 1, GpuType: ""}, Percentage: 8.43 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 16000, MilliGpu: 678, GpuNumber: 1, GpuType: "T4"}, Percentage: 8.36 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 8000, MilliGpu: 500, GpuNumber: 1, GpuType: ""}, Percentage: 8.29 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 6000, MilliGpu: 511, GpuNumber: 1, GpuType: ""}, Percentage: 8.11 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 14000, MilliGpu: 1000, GpuNumber: 2, GpuType: "2080"}, Percentage: 0.54 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 4000, MilliGpu: 1000, GpuNumber: 1, GpuType: "2080"}, Percentage: 0.43 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 32000, MilliGpu: 1000, GpuNumber: 2, GpuType: "T4"}, Percentage: 0.43 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 16000, MilliGpu: 1000, GpuNumber: 1, GpuType: "V100M16"}, Percentage: 0.40 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 64000, MilliGpu: 1000, GpuNumber: 2, GpuType: ""}, Percentage: 0.40 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 10000, MilliGpu: 1000, GpuNumber: 2, GpuType: ""}, Percentage: 0.40 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 11400, MilliGpu: 1000, GpuNumber: 1, GpuType: "T4"}, Percentage: 0.36 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 16000, MilliGpu: 1000, GpuNumber: 1, GpuType: "T4"}, Percentage: 0.36 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 4000, MilliGpu: 1000, GpuNumber: 2, GpuType: ""}, Percentage: 0.36 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 14000, MilliGpu: 1000, GpuNumber: 2, GpuType: "V100M16"}, Percentage: 0.36 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 8000, MilliGpu: 1000, GpuNumber: 4, GpuType: ""}, Percentage: 0.36 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 16000, MilliGpu: 1000, GpuNumber: 2, GpuType: ""}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 2000, MilliGpu: 1000, GpuNumber: 1, GpuType: "T4"}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 6000, MilliGpu: 1000, GpuNumber: 1, GpuType: ""}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 4000, MilliGpu: 1000, GpuNumber: 1, GpuType: ""}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 5000, MilliGpu: 1000, GpuNumber: 1, GpuType: ""}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 32000, MilliGpu: 1000, GpuNumber: 4, GpuType: "V100M16"}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 32000, MilliGpu: 1000, GpuNumber: 2, GpuType: ""}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 24000, MilliGpu: 1000, GpuNumber: 8, GpuType: "2080"}, Percentage: 0.32 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 40000, MilliGpu: 1000, GpuNumber: 4, GpuType: ""}, Percentage: 0.29 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 32000, MilliGpu: 1000, GpuNumber: 8, GpuType: ""}, Percentage: 0.29 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 32000, MilliGpu: 1000, GpuNumber: 1, GpuType: "T4"}, Percentage: 0.29 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 16000, MilliGpu: 1000, GpuNumber: 1, GpuType: ""}, Percentage: 0.25 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 7000, MilliGpu: 1000, GpuNumber: 1, GpuType: "V100M16"}, Percentage: 0.25 / 100})
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 24000, MilliGpu: 1000, GpuNumber: 1, GpuType: "T4"}, Percentage: 0.25 / 100})
	return typicalPods
}

func TestNodeGpuFragAmountBellman_EightGpu(t *testing.T) {
	nodeRes := simontype.NodeResource{"node_2C_4x1080", 78000,
		[]int64{1000, 1000, 1000, 1000, 1000, 1000, 535, 70}, 8, "V100M32"} // GpuType: See MapGpuTypeMemoryMiB
	typicalPods := TestingGenerateGetTypicalPods()
	dp := sync.Map{}
	frag := NodeGpuFragBellman(nodeRes, typicalPods, &dp, 1.0)
	//assert.Equal(t, FragAmount{"node_2C_4x1080", []float64{211.014, 205.056, 730.296, 870.534, 0, 0, 0}}.Repr(), fragAmount.Repr())
	assert.InDelta(t, 160.73, frag, 0.01)
	// Q1LackBoth: 0, Q2LackGpu: 1, Q3Satisfied: 2, Q4LackCpu: 3, XLSatisfied: 4, XRLackCPU: 5, NoAccess: 6
}

func TestNodeGpuShareFragAmountScore(t *testing.T) {
	typicalPods := TestingGenerateGetTypicalPods()
	nodeRes := simontype.NodeResource{"4x1080_used", 1000, []int64{200, 1000, 1000, 500}, 4, "1080"}
	score := NodeGpuShareFragAmountScore(nodeRes, typicalPods)
	assert.InDelta(t, 2566.62, score, 0.01)

	nodeRes = simontype.NodeResource{"4x1080_full", 1000, []int64{1000, 1000, 1000, 1000}, 4, "1080"}
	score = NodeGpuShareFragAmountScore(nodeRes, typicalPods)
	assert.InDelta(t, 3802.40, score, 0.01)

	nodeRes = simontype.NodeResource{"8x1080_full", 1000, []int64{1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000}, 8, "1080"}
	score = NodeGpuShareFragAmountScore(nodeRes, typicalPods)
	assert.InDelta(t, 7604.80, score, 0.01)

	typicalPods = simontype.TargetPodList{}
	typicalPods = append(typicalPods, simontype.TargetPod{TargetPodResource: simontype.PodResource{MilliCpu: 6000, MilliGpu: 465, GpuNumber: 1, GpuType: ""}, Percentage: 9.33 / 100})
	nodeRes = simontype.NodeResource{"4x1080_used_lack_CPU", 1000, []int64{200, 1000, 1000, 500}, 4, "1080"}
	assert.Equal(t, GetNodePodFrag(nodeRes, typicalPods[0].TargetPodResource), Q4LackCpu)
	assert.Equal(t, int64(2700), GetGpuMilliLeftTotal(nodeRes))
	score = NodeGpuShareFragAmountScore(nodeRes, typicalPods)
	assert.InDelta(t, 251.91, score, 0.01)
}

func TestGetGpuFragMilliByNodeResAndPodRes(t *testing.T) {
	nodeRes := simontype.NodeResource{"4x1080_used", 1000, []int64{200, 1000, 1000, 500}, 4, "1080"}
	podRes := simontype.PodResource{100, 1000, 2, "1080"}
	fragMilli := GetGpuFragMilliByNodeResAndPodRes(nodeRes, podRes)
	assert.Equal(t, int64(700), fragMilli)

	nodeRes = simontype.NodeResource{"4x1080_full", 1000, []int64{1000, 1000, 1000, 1000}, 4, "1080"}
	podRes = simontype.PodResource{100, 1000, 2, "1080"}
	fragMilli = GetGpuFragMilliByNodeResAndPodRes(nodeRes, podRes)
	assert.Equal(t, int64(0), fragMilli)

	nodeRes = simontype.NodeResource{"8x1080_full", 1000, []int64{1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000}, 8, "1080"}
	podRes = simontype.PodResource{100, 1000, 2, "1080"}
	fragMilli = GetGpuFragMilliByNodeResAndPodRes(nodeRes, podRes)
	assert.Equal(t, int64(0), fragMilli)

	nodeRes = simontype.NodeResource{"4x1080_used", 1000, []int64{200, 1000, 1000, 500}, 4, "1080"}
	podRes = simontype.PodResource{100, 200, 2, "1080"}
	fragMilli = GetGpuFragMilliByNodeResAndPodRes(nodeRes, podRes)
	assert.Equal(t, int64(0), fragMilli)
}

/*
func TestNodeGpuFragAmountBellman_Direct(t *testing.T) {
	nodeRes := simontype.NodeResource{"node_1C_4x1080", 1000,
		[]int64{200, 600, 350, 100}, 4, "1080"} // GpuType: See MapGpuTypeMemoryMiB
	typicalPods := simontype.TargetPodList{}
	pod := simontype.TargetPod{}

	// Q1LackBoth
	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 2000, MilliGpu: 1000, GpuNumber: 4, GpuType: "",
	}, Percentage: 0.1}
	typicalPods = append(typicalPods, pod)

	// Q2LackGpu
	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1000, MilliGpu: 1000, GpuNumber: 4, GpuType: "",
	}, Percentage: 0.2}
	typicalPods = append(typicalPods, pod)

	// Q3Satisfied
	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1000, MilliGpu: 100, GpuNumber: 4, GpuType: "",
	}, Percentage: 0.3}
	typicalPods = append(typicalPods, pod)

	// Q4LackCpu
	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 2000, MilliGpu: 100, GpuNumber: 1, GpuType: "",
	}, Percentage: 0.1}
	typicalPods = append(typicalPods, pod)

	// XLSatisfied
	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1000, MilliGpu: 0, GpuNumber: 0, GpuType: "",
	}, Percentage: 0.1}
	typicalPods = append(typicalPods, pod)

	// XRLackCPU
	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 2000, MilliGpu: 0, GpuNumber: 0, GpuType: "",
	}, Percentage: 0.1}
	typicalPods = append(typicalPods, pod)

	// NoAccess
	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1000, MilliGpu: 200, GpuNumber: 2, GpuType: "V100M32",
	}, Percentage: 0.1}
	typicalPods = append(typicalPods, pod)

	dp := sync.Map{}
	fragAmountBellman := NodeGpuFragAmountBellman(nodeRes, typicalPods, &dp)
	//fragAmount := NodeGpuFragAmount(nodeRes, typicalPods)
	frag := NodeGpuFragBellman(nodeRes, typicalPods, &dp, 1.0)
	//assert.Equal(t, FragAmount{"node_1C_4x1080", []float64{125, 250, 375, 125, 125, 125, 125}}, fragAmount)
	//assert.Equal(t, fragAmount, fragAmountBellman)
	assert.InDelta(t, fragAmountBellman.FragAmountSumExceptQ3(), frag, 0.01)
	// Q1LackBoth: 0, Q2LackGpu: 1, Q3Satisfied: 2, Q4LackCpu: 3, XLSatisfied: 4, XRLackCPU: 5, NoAccess: 6
}

func TestNodeGpuFragAmountBellman_NonDirect(t *testing.T) {
	nodeRes := simontype.NodeResource{"node_2C_4x1080", 2000,
		[]int64{1000, 1000, 1000, 1000}, 4, "V100M32"} // GpuType: See MapGpuTypeMemoryMiB
	typicalPods := simontype.TargetPodList{}
	pod := simontype.TargetPod{}

	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1500, MilliGpu: 900, GpuNumber: 4, GpuType: "",
	}, Percentage: 0.1}
	typicalPods = append(typicalPods, pod)

	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1000, MilliGpu: 900, GpuNumber: 4, GpuType: "",
	}, Percentage: 0.2}
	typicalPods = append(typicalPods, pod)

	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 500, MilliGpu: 100, GpuNumber: 4, GpuType: "",
	}, Percentage: 0.3}
	typicalPods = append(typicalPods, pod)

	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1500, MilliGpu: 100, GpuNumber: 1, GpuType: "",
	}, Percentage: 0.2}
	typicalPods = append(typicalPods, pod)

	pod = simontype.TargetPod{TargetPodResource: simontype.PodResource{
		MilliCpu: 1000, MilliGpu: 200, GpuNumber: 2, GpuType: "1080",
	}, Percentage: 0.2}
	typicalPods = append(typicalPods, pod)

	dp := sync.Map{}
	fragAmount := NodeGpuFragAmountBellman(nodeRes, typicalPods, &dp)
	frag := NodeGpuFragBellman(nodeRes, typicalPods, &dp, 1.0)
	//assert.Equal(t, FragAmount{"node_2C_4x1080", []float64{211.014, 205.056, 730.296, 870.534, 0, 0, 0}}.Repr(), fragAmount.Repr())
	assert.InDelta(t, fragAmount.FragAmountSumExceptQ3(), frag, 0.01)
	// Q1LackBoth: 0, Q2LackGpu: 1, Q3Satisfied: 2, Q4LackCpu: 3, XLSatisfied: 4, XRLackCPU: 5, NoAccess: 6
}

*/
