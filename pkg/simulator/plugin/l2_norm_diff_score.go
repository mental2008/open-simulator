package plugin

import (
	"context"
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	simontype "github.com/alibaba/open-simulator/pkg/type"
	"github.com/alibaba/open-simulator/pkg/utils"
)

type L2NormDiffScorePlugin struct {
	handle framework.Handle
}

var _ framework.ScorePlugin = &L2NormDiffScorePlugin{}

func NewL2NormDiffScorePlugin(configuration runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &L2NormDiffScorePlugin{
		handle: handle,
	}, nil
}

func (plugin *L2NormDiffScorePlugin) Name() string {
	return simontype.L2NormDiffScorePluginName
}

func (plugin *L2NormDiffScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) (int64, *framework.Status) {
	// < common procedure that prepares node, podRes, nodeRes>
	node, err := plugin.handle.ClientSet().CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("failed to get node %s: %s\n", nodeName, err.Error()))
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
	// </common procedure that prepares node, podRes, nodeRes>

	nodeCap := utils.GetNodeAllocatableCpuGpu(node)
	nodeVec := utils.NormalizeVector(nodeRes.ToResourceVec(), nodeCap)
	podVec := utils.NormalizeVector(podRes.ToResourceVec(), nodeCap)

	score := utils.CalculateL2NormDiff(nodeVec, podVec)
	log.Tracef("L2 Norm Diff score between nodeRes(%s) and podRes(%s) with nodeCap(%v): %.4f\n",
		nodeRes.Repr(), podRes.Repr(), nodeCap, score)
	if score == -1 {
		return framework.MinNodeScore, framework.NewStatus(framework.Success)
	}
	score /= float64(len(podVec)) // normalize score to [0, 1]
	score = 1 - score             // the larger the norm diff, the lower the score
	return int64(math.Round(float64(framework.MaxNodeScore) * score)), framework.NewStatus(framework.Success)
}

func (plugin *L2NormDiffScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}
