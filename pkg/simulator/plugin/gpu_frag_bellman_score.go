package plugin

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	simontype "github.com/alibaba/open-simulator/pkg/type"
	"github.com/alibaba/open-simulator/pkg/utils"
)

// GpuFragBellmanScorePlugin is a plugin for scheduling framework, scoring pods by GPU fragmentation amount
type GpuFragBellmanScorePlugin struct {
	handle        framework.Handle
	typicalPods   *simontype.TargetPodList
	fragRatioMemo *sync.Map
	sync.RWMutex
}

// Just to check whether the implemented struct fits the interface
var _ framework.ScorePlugin = &GpuFragBellmanScorePlugin{}

func NewGpuFragBellmanScorePlugin(configuration runtime.Object, handle framework.Handle, typicalPods *simontype.TargetPodList, fragRatioMemo *sync.Map) (framework.Plugin, error) {
	gpuFragScorePlugin := &GpuFragBellmanScorePlugin{
		handle:        handle,
		typicalPods:   typicalPods,
		fragRatioMemo: fragRatioMemo,
	}
	return gpuFragScorePlugin, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (plugin *GpuFragBellmanScorePlugin) Name() string {
	return simontype.GpuFragBellmanScorePluginName
}

// Score invoked at the score extension point.
func (plugin *GpuFragBellmanScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) (int64, *framework.Status) {
	// < common procedure that prepares podRes, nodeRes, newNodeRes for Frag related score plugins>
	if podReq, _ := resourcehelper.PodRequestsAndLimits(pod); len(podReq) == 0 {
		return framework.MaxNodeScore, framework.NewStatus(framework.Success)
	}

	nodeResPtr := utils.GetNodeResourceViaHandleAndName(plugin.handle, nodeName)
	if nodeResPtr == nil {
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("failed to get nodeRes(%s)\n", nodeName))
	}
	nodeRes := *nodeResPtr

	podRes := utils.GetPodResource(pod)
	if !utils.IsNodeAccessibleToPod(nodeRes, podRes) {
		log.Errorf("Node (%s) %s does not match GPU type request of pod %s. Should be filtered by GpuSharePlugin", nodeName, nodeRes.Repr(), podRes.Repr())
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("Node (%s) %s does not match GPU type request of pod %s\n", nodeName, nodeRes.Repr(), podRes.Repr()))
	}

	newNodeRes, err := nodeRes.Sub(podRes)
	if err != nil {
		log.Errorf(err.Error())
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("Node (%s) %s does not have sufficient resource for pod (%s) %s\n", nodeName, nodeRes.Repr(), pod.Name, podRes.Repr()))
	}

	if plugin.typicalPods == nil {
		log.Errorf("typical pods list is empty\n")
		return framework.MinNodeScore, framework.NewStatus(framework.Error, "typical pods list is empty\n")
	}
	// </common procedure that prepares podRes, nodeRes, newNodeRes for Frag related score plugins>

	// Key difference between Gpu-Frag-Score and Gpu-Frag-Score-Bellman:
	///*
	plugin.Lock()
	defer plugin.Unlock()
	nodeGpuFragValue := utils.NodeGpuFragBellman(nodeRes, *plugin.typicalPods, plugin.fragRatioMemo, 1.0)
	newNodeGpuFragValue := utils.NodeGpuFragBellman(newNodeRes, *plugin.typicalPods, plugin.fragRatioMemo, 1.0)
	//*/
	//nodeGpuFragValue := utils.NodeGpuFragBellmanEarlyStop(nodeRes, *plugin.typicalPods, plugin.fragRatioMemo)
	//newNodeGpuFragValue := utils.NodeGpuFragBellmanEarlyStop(newNodeRes, *plugin.typicalPods, plugin.fragRatioMemo)

	score := int64(nodeGpuFragValue - newNodeGpuFragValue) // The higher, the better. Negative means fragment amount increases, which is among the worst cases.
	return score, framework.NewStatus(framework.Success)
}

// ScoreExtensions of the Score plugin.
func (plugin *GpuFragBellmanScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return plugin
}

// NormalizeScore invoked after scoring all nodes.
func (plugin *GpuFragBellmanScorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, scores framework.NodeScoreList) *framework.Status {
	return NormalizeScore(scores)
}
