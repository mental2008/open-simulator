package plugin

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/pquerna/ffjson/ffjson"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	externalclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/alibaba/open-simulator/pkg/algo"
	simontype "github.com/alibaba/open-simulator/pkg/type"
	gpusharecache "github.com/alibaba/open-simulator/pkg/type/open-gpu-share/cache"
	gpushareutils "github.com/alibaba/open-simulator/pkg/type/open-gpu-share/utils"
)

// GpuSharePlugin is a plugin for scheduling framework
type GpuSharePlugin struct {
	sync.RWMutex
	fakeclient          externalclientset.Interface
	cache               *gpusharecache.SchedulerCache
	podToUpdateCacheMap map[string]*corev1.Pod // key: getPodMapKey(): return pod.Namespace+pod.Name
}

// Just to check whether the implemented struct fits the interface
var _ framework.FilterPlugin = &GpuSharePlugin{}
var _ framework.ScorePlugin = &GpuSharePlugin{}
var _ framework.ReservePlugin = &GpuSharePlugin{}
var _ framework.BindPlugin = &GpuSharePlugin{}

func NewGpuSharePlugin(fakeclient externalclientset.Interface, configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	gpuSharePlugin := &GpuSharePlugin{
		fakeclient:          fakeclient,
		podToUpdateCacheMap: make(map[string]*corev1.Pod),
	}
	gpuSharePlugin.InitSchedulerCache()
	f.SharedInformerFactory().Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				if pod, ok := obj.(*corev1.Pod); ok {
					if gpushareutils.GetGpuMilliFromPodAnnotation(pod) > 0 {
						//namespace, name := pod.Namespace, pod.Name
						//fmt.Printf("delete_gpu_bgn: pod %s/%s\n", namespace, name)
						_ = gpuSharePlugin.removePod(pod)
						//fmt.Printf("delete_gpu_end: pod %s/%s\n", namespace, name)
					}
				}
			}})
	return gpuSharePlugin, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (plugin *GpuSharePlugin) Name() string {
	return simontype.OpenGpuSharePluginName
}

// Filter Plugin
// Filter filters out non-allocatable nodes
func (plugin *GpuSharePlugin) Filter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	//fmt.Printf("filter_gpu: pod %s/%s, nodeName %s\n", pod.Namespace, pod.Name, nodeInfo.Node().Name)
	// check if the pod requires GPU resources
	podGpuMem := gpushareutils.GetGpuMilliFromPodAnnotation(pod)
	if podGpuMem <= 0 {
		// the node is schedulable if pod does not require GPU resources
		//klog.Infof("[Filter] Pod: %v/%v, podGpuMem <= 0: %v", pod.GetNamespace(), pod.GetName(), podGpuMem)
		return framework.NewStatus(framework.Success)
	}
	//klog.Infof("[Filter] Pod: %v/%v, podGpuMem: %v", pod.GetNamespace(), pod.GetName(), podGpuMem)

	// check if the node have GPU resources
	node := nodeInfo.Node()
	nodeGpuMem := gpushareutils.GetTotalGpuMemory(node)
	if nodeGpuMem < podGpuMem {
		//klog.Infof("[Filter] Unschedulable, Node: %v, nodeGpuMem: %v", node.GetName(), nodeGpuMem)
		return framework.NewStatus(framework.Unschedulable, "Node:"+nodeInfo.Node().Name)
	}
	//klog.Infof("[Filter] Schedulable, Node: %v, nodeGpuMem: %v", node.GetName(), nodeGpuMem)

	// check if any of the GPU has such resources
	gpuNodeInfo, err := plugin.cache.GetGpuNodeInfo(node.Name)
	if err != nil {
		return framework.NewStatus(framework.Unschedulable, "Node:"+nodeInfo.Node().Name)
	}
	_, found := gpuNodeInfo.AllocateGpuId(pod)
	if !found {
		return framework.NewStatus(framework.Unschedulable, "Node:"+nodeInfo.Node().Name)
	}

	return framework.NewStatus(framework.Success)
}

// Score Plugin
// Score invoked at the score extension point.
func (plugin *GpuSharePlugin) Score(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) (int64, *framework.Status) {
	//fmt.Printf("score_gpu: pod %s/%s, nodeName %s\n", pod.Namespace, pod.Name, nodeName)
	podReq, _ := resourcehelper.PodRequestsAndLimits(pod)
	if len(podReq) == 0 {
		return framework.MaxNodeScore, framework.NewStatus(framework.Success)
	}

	node, err := plugin.fakeclient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return int64(framework.MinNodeScore), framework.NewStatus(framework.Error, fmt.Sprintf("failed to get node %s: %s\n", nodeName, err.Error()))
	}

	res := float64(0)
	for resourceName := range node.Status.Allocatable {
		podAllocatedRes := podReq[resourceName]
		nodeAvailableRes := node.Status.Allocatable[resourceName]
		nodeAvailableRes.Sub(podAllocatedRes)
		share := algo.Share(podAllocatedRes.AsApproximateFloat64(), nodeAvailableRes.AsApproximateFloat64())
		if share > res {
			res = share
		}
	}

	score := int64(float64(framework.MaxNodeScore-framework.MinNodeScore) * res)
	//klog.Infof("[Score] Pod: %v at Node: %v => Score: %d", pod.Name, nodeName, score)
	return score, framework.NewStatus(framework.Success)
}

// ScoreExtensions of the Score plugin.
func (plugin *GpuSharePlugin) ScoreExtensions() framework.ScoreExtensions {
	return plugin // if there is no NormalizeScore, return nil.
}

// NormalizeScore invoked after scoring all nodes.
func (plugin *GpuSharePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, scores framework.NodeScoreList) *framework.Status {
	// Find highest and lowest scores.
	var highest int64 = -math.MaxInt64
	var lowest int64 = math.MaxInt64
	for _, nodeScore := range scores {
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
	}

	// Transform the highest to lowest score range to fit the framework's min to max node score range.
	oldRange := highest - lowest
	newRange := framework.MaxNodeScore - framework.MinNodeScore
	for i, nodeScore := range scores {
		if oldRange == 0 {
			scores[i].Score = framework.MinNodeScore
		} else {
			scores[i].Score = ((nodeScore.Score - lowest) * newRange / oldRange) + framework.MinNodeScore
		}
	}

	return framework.NewStatus(framework.Success)
}

func (plugin *GpuSharePlugin) updateNode(node *corev1.Node) error {
	nodeGpuInfoStr, err := plugin.ExportGpuNodeInfoAsNodeGpuInfo(node.Name)
	if err != nil {
		return err
	}
	if data, err := ffjson.Marshal(nodeGpuInfoStr); err != nil {
		return err
	} else {
		metav1.SetMetaDataAnnotation(&node.ObjectMeta, simontype.AnnoNodeGpuShare, string(data))
	}
	//fmt.Printf("updateNode: %v with anno: %s\n", nodeGpuInfo, node.ObjectMeta.Annotations)

	if _, err := plugin.fakeclient.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to Update node %s", node.Name)
	}
	return nil
}

func (plugin *GpuSharePlugin) addOrUpdatePod(pod *corev1.Pod) error {
	if err := plugin.cache.AddOrUpdatePod(pod); err != nil { // requires pod.Spec.NodeName specified
		return err
	}
	if pod.Spec.NodeName == "" {
		return fmt.Errorf("pod unscheduled: %s/%s", pod.Namespace, pod.Name)
	}
	node, err := plugin.fakeclient.CoreV1().Nodes().Get(context.Background(), pod.Spec.NodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	//fmt.Printf("addOrUpdatePod: %s\n", pod.Name)
	if err = plugin.updateNode(node); err != nil {
		return err
	}
	return nil
}

func (plugin *GpuSharePlugin) removePod(pod *corev1.Pod) error {
	nodeName := pod.Spec.NodeName
	if nodeName == "" {
		return fmt.Errorf("pod not scheduled when removed: %s/%s", pod.Namespace, pod.Name)
	}
	node, err := plugin.fakeclient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	plugin.cache.RemovePod(pod)
	//fmt.Printf("removePod: %s\n", pod.Name)
	if err = plugin.updateNode(node); err != nil {
		return err
	}
	return nil
}

// Reserve Plugin
// Reserve updates the GPU resource of the given node, according to the pod's request.
func (plugin *GpuSharePlugin) Reserve(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) *framework.Status {
	//fmt.Printf("reserve_gpu: pod %s/%s, nodeName %s\n", pod.Namespace, pod.Name, nodeName)
	if gpushareutils.GetGpuMilliFromPodAnnotation(pod) <= 0 {
		return framework.NewStatus(framework.Success) // non-GPU pods are skipped
	}
	plugin.Lock()
	defer plugin.Unlock()

	// get PodCopy but NOT update it
	podCopy, err := plugin.MakePodCopyReadyForBindUpdate(pod, nodeName)
	if err != nil {
		klog.Errorf("The node %s can't place the pod %s in ns %s,and the pod spec is %v. err: %s", pod.Spec.NodeName, pod.Name, pod.Namespace, pod, err)
		return framework.NewStatus(framework.Error, err.Error())
	}
	plugin.podToUpdateCacheMap[getPodMapKey(pod)] = podCopy

	// get node from fakeclient and update Node
	if err = plugin.addOrUpdatePod(podCopy); err != nil {
		//fmt.Printf("addOrUpdatePod: pod %s/%s, nodeName %s, error %v\n", pod.Namespace, pod.Name, nodeName, err)
		return framework.NewStatus(framework.Error, err.Error())
	}

	return framework.NewStatus(framework.Success)
}

// Unreserve undoes the GPU resource updated in Reserve function.
func (plugin *GpuSharePlugin) Unreserve(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) {
	plugin.Lock()
	defer plugin.Unlock()

	panic(fmt.Errorf("unreserve would lead to extra pod/node Updates, causing sim.updateBarrier not equal"))
	//if err := plugin.removePod(pod); err != nil {
	//	klog.Errorf(err.Error())
	//}
}

// Bind Plugin
// Bind updates the GPU resources of the pod.
func (plugin *GpuSharePlugin) Bind(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) *framework.Status {
	//fmt.Printf("bind_gpu: pod %s/%s, nodeName %s\n", pod.Namespace, pod.Name, nodeName)
	if gpushareutils.GetGpuMilliFromPodAnnotation(pod) <= 0 {
		return framework.NewStatus(framework.Skip) // non-GPU pods are skipped
	}

	podCopy, ok := plugin.podToUpdateCacheMap[getPodMapKey(pod)]
	if !ok {
		klog.Errorf("No podToUpdate found, which should not happen since it should have failed in ReservePlugin")
		return framework.NewStatus(framework.Error, fmt.Sprintf("No podToUpdate found"))
	}
	_, err := plugin.fakeclient.CoreV1().Pods(podCopy.Namespace).Update(context.TODO(), podCopy, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("fake update error %v", err)
		return framework.NewStatus(framework.Error, fmt.Sprintf("Unable to add new pod: %v", err))
	}
	delete(plugin.podToUpdateCacheMap, getPodMapKey(pod)) // avoid memory leakage
	//klog.Infof("Allocate() ---- pod %s in ns %s is allocated to node %s ----", podCopy.Name, podCopy.Namespace, podCopy.Spec.NodeName)
	return nil
}

// Util Functions

func (plugin *GpuSharePlugin) ExportGpuNodeInfoAsNodeGpuInfo(nodeName string) (*gpusharecache.GpuNodeInfoStr, error) {
	if gpuNodeInfo, err := plugin.cache.GetGpuNodeInfo(nodeName); err != nil {
		return nil, err
	} else {
		nodeGpuInfoStr := gpuNodeInfo.ExportGpuNodeInfoAsStr()
		return nodeGpuInfoStr, nil
	}
}

func (plugin *GpuSharePlugin) NodeGet(name string) (*corev1.Node, error) {
	return plugin.fakeclient.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
}

func (plugin *GpuSharePlugin) PodGet(name string, namespace string) (*corev1.Pod, error) {
	return plugin.fakeclient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (plugin *GpuSharePlugin) InitSchedulerCache() {
	plugin.cache = gpusharecache.NewSchedulerCache(plugin) // here `plugin` implements the NodePodGetter interface
}

func (plugin *GpuSharePlugin) MakePodCopyReadyForBindUpdate(pod *corev1.Pod, nodeName string) (*corev1.Pod, error) {
	gpuNodeInfo, err := plugin.cache.GetGpuNodeInfo(nodeName)
	if err != nil {
		return nil, err
	}

	devId, found := gpuNodeInfo.AllocateGpuId(pod)
	if !found {
		err := fmt.Errorf("cannot find a GPU to allocate pod %s at node %s", getPodMapKey(pod), nodeName)
		return nil, err
	}

	podCopy := gpushareutils.GetUpdatedPodAnnotationSpec(pod, devId)
	podCopy.Spec.NodeName = nodeName
	podCopy.Status.Phase = corev1.PodRunning
	return podCopy, nil
}

func getPodMapKey(pod *corev1.Pod) string {
	return pod.Namespace + pod.Name
}
