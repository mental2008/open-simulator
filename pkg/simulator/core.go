package simulator

import (
	"fmt"
	"math/rand"
	"os"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/alibaba/open-simulator/pkg/api/v1alpha1"
	"github.com/alibaba/open-simulator/pkg/type"
	"github.com/alibaba/open-simulator/pkg/utils"
)

type ResourceTypes struct {
	Nodes                  []*corev1.Node
	Pods                   []*corev1.Pod
	DaemonSets             []*appsv1.DaemonSet
	StatefulSets           []*appsv1.StatefulSet
	Deployments            []*appsv1.Deployment
	ReplicationControllers []*corev1.ReplicationController
	ReplicaSets            []*appsv1.ReplicaSet
	Services               []*corev1.Service
	PersistentVolumeClaims []*corev1.PersistentVolumeClaim
	StorageClasss          []*storagev1.StorageClass
	PodDisruptionBudgets   []*policyv1beta1.PodDisruptionBudget
	Jobs                   []*batchv1.Job
	CronJobs               []*batchv1beta1.CronJob
}

type AppResource struct {
	Name     string
	Resource ResourceTypes
}

type Interface interface {
	RunCluster(cluster ResourceTypes) ([]simontype.UnscheduledPod, error)
	ScheduleApp(AppResource) ([]simontype.UnscheduledPod, error)
	SchedulePods(pods []*corev1.Pod) []simontype.UnscheduledPod

	ClusterAnalysis(tag string) (utils.FragAmount, []utils.ResourceSummary)
	ClusterGpuFragReport()
	GetClusterNodeStatus() []simontype.NodeStatus

	SetWorkloadPods(pods []*corev1.Pod)
	SetTypicalPods()
	SetSkylinePods()
	RecordPodTotalResourceReq(pods []*corev1.Pod) (int64, int64)
	RecordNodeTotalResource(nodes []*corev1.Node) (int64, int64)
	TunePodsByNodeTotalResource(pods []*corev1.Pod, config v1alpha1.WorkloadTuningConfig) []*corev1.Pod

	ExportPodSnapshotInYaml(unschedulePods []simontype.UnscheduledPod, filePath string)
	ExportNodeSnapshotInCSV(filePath string)
	ExportPodSnapshotInCSV(filePath string)

	SortClusterPods(pods []*corev1.Pod)

	RunWorkloadInflationEvaluation(tag string)

	GetCustomConfig() v1alpha1.CustomConfig

	DescheduleCluster() []simontype.UnscheduledPod

	Close()
}

// Simulate
// 参数
// 1. 由使用方自己生成 cluster 和 apps 传参
// 2. apps 将按照顺序模拟部署
// 3. 存储信息以 Json 形式填入对应的 Node 资源中
// 返回值
// 1. error 不为空表示函数执行失败
// 2. error 为空表示函数执行成功，通过 SimulateResult 信息获取集群模拟信息。其中 UnscheduledPods 表示无法调度的 Pods，若其为空表示模拟调度成功；NodeStatus 会详细记录每个 Node 上的 Pod 情况。
func Simulate(cluster ResourceTypes, apps []AppResource, opts ...Option) (*simontype.SimulateResult, error) {
	// init simulator
	sim, err := New(opts...)
	if err != nil {
		return nil, err
	}
	defer sim.Close()

	cluster.Pods, err = GetValidPodExcludeDaemonSet(cluster)
	if err != nil {
		return nil, err
	}
	log.Infof("Number of original workload pods: %d", len(cluster.Pods))
	sim.SetWorkloadPods(cluster.Pods)
	sim.SetTypicalPods()
	sim.SetSkylinePods()
	sim.ClusterGpuFragReport()

	customConfig := sim.GetCustomConfig()
	rand.Seed(customConfig.WorkloadTuningConfig.Seed)
	log.Debugf("Random Seed: %d, Random Int: %d", customConfig.WorkloadTuningConfig.Seed, rand.Int())
	for _, item := range cluster.DaemonSets {
		validPods, err := utils.MakeValidPodsByDaemonset(item, cluster.Nodes)
		if err != nil {
			return nil, err
		}
		cluster.Pods = append(cluster.Pods, validPods...)
	}

	var failedPods []simontype.UnscheduledPod

	// run cluster
	sim.SortClusterPods(cluster.Pods)
	sim.RecordPodTotalResourceReq(cluster.Pods)
	sim.RecordNodeTotalResource(cluster.Nodes)

	if customConfig.WorkloadTuningConfig.Ratio > 0 {
		// <= 0 means no tuning, keeping the cluster.Pods == sim.workloadPods
		cluster.Pods = sim.TunePodsByNodeTotalResource(cluster.Pods, customConfig.WorkloadTuningConfig)
	}

	unscheduledPods, err := sim.RunCluster(cluster) // Existing pods in the cluster are scheduled here.
	if err != nil {
		return nil, err
	}
	failedPods = append(failedPods, unscheduledPods...)
	utils.ReportFailedPods(failedPods)
	sim.ClusterAnalysis(TagInitSchedule)

	if customConfig.ExportConfig.PodSnapshotYamlFilePrefix != "" {
		// filePath: prefix/InitSchedule/pod-snapshot.yaml
		prefix := customConfig.ExportConfig.PodSnapshotYamlFilePrefix
		fileDir := fmt.Sprintf("%s/%s", prefix, TagInitSchedule)
		if e := os.MkdirAll(fileDir, os.FileMode(0777)); e != nil {
			log.Errorf("MkdirAll(%s, 0777) failed: %s", fileDir, e.Error())
		} else {
			filePath := fmt.Sprintf("%s/%s", fileDir, "pod-snapshot.yaml")
			sim.ExportPodSnapshotInYaml(unscheduledPods, filePath)
		}
	}
	if customConfig.ExportConfig.NodeSnapshotCSVFilePrefix != "" {
		// filePath: prefix/InitSchedule/node-snapshot.csv
		prefix := customConfig.ExportConfig.NodeSnapshotCSVFilePrefix
		fileDir := fmt.Sprintf("%s/%s", prefix, TagInitSchedule)
		if e := os.MkdirAll(fileDir, os.FileMode(0777)); e != nil {
			log.Errorf("MkdirAll(%s, 0777) failed: %s", fileDir, e.Error())
		} else {
			filePath := fmt.Sprintf("%s/%s", fileDir, "node-snapshot.csv")
			sim.ExportNodeSnapshotInCSV(filePath)
			podFilePath := fmt.Sprintf("%s/%s", fileDir, "pod-snapshot.csv")
			sim.ExportPodSnapshotInCSV(podFilePath)
		}
	}

	if customConfig.WorkloadInflationConfig.Ratio > 1 {
		sim.RunWorkloadInflationEvaluation(TagScheduleInflation)
	}

	if customConfig.NewWorkloadConfig != "" {
		resources, err := CreateClusterResourceFromClusterConfig(customConfig.NewWorkloadConfig)
		if err != nil {
			return nil, err
		}
		newWorkloadPods, err := GetValidPodExcludeDaemonSet(resources)
		if err != nil {
			return nil, err
		}
		log.Infof("Number of new workload pods: %d\n", len(newWorkloadPods))
		sim.SetWorkloadPods(newWorkloadPods)
		sim.SetTypicalPods()
		sim.ClusterGpuFragReport()
	}
	if customConfig.DescheduleConfig.Policy != "" {
		unscheduledPods = sim.DescheduleCluster()
		failedPods = append(failedPods, unscheduledPods...)
		sim.ClusterAnalysis(TagPostDeschedule)
		sim.ClusterGpuFragReport()

		if customConfig.ExportConfig.PodSnapshotYamlFilePrefix != "" {
			// filePath: prefix/PostDeschedule/pod-snapshot.yaml
			prefix := customConfig.ExportConfig.PodSnapshotYamlFilePrefix
			fileDir := fmt.Sprintf("%s/%s", prefix, TagPostDeschedule)
			if e := os.MkdirAll(fileDir, os.FileMode(0777)); e != nil {
				log.Errorf("MkdirAll(%s, 0777) failed: %s", fileDir, e.Error())
			} else {
				filePath := fmt.Sprintf("%s/%s", fileDir, "pod-snapshot.yaml")
				sim.ExportPodSnapshotInYaml(unscheduledPods, filePath)
			}
		}
		if customConfig.ExportConfig.NodeSnapshotCSVFilePrefix != "" {
			// filePath: prefix/PostDeschedule/node-snapshot.csv
			prefix := customConfig.ExportConfig.NodeSnapshotCSVFilePrefix
			fileDir := fmt.Sprintf("%s/%s", prefix, TagPostDeschedule)
			if e := os.MkdirAll(fileDir, os.FileMode(0777)); e != nil {
				log.Errorf("MkdirAll(%s, 0777) failed: %s", fileDir, e.Error())
			} else {
				filePath := fmt.Sprintf("%s/%s", fileDir, "node-snapshot.csv")
				sim.ExportNodeSnapshotInCSV(filePath)
				podFilePath := fmt.Sprintf("%s/%s", fileDir, "pod-snapshot.csv")
				sim.ExportNodeSnapshotInCSV(podFilePath)
			}
		}
	}
	if customConfig.NewWorkloadConfig != "" || customConfig.DescheduleConfig.Policy != "" {
		if customConfig.WorkloadInflationConfig.Ratio > 1 {
			sim.RunWorkloadInflationEvaluation(TagDescheduleInflation)
		}
	}

	// schedule pods
	for _, app := range apps {
		unscheduledPods, err = sim.ScheduleApp(app)
		if err != nil {
			return nil, err
		}
		failedPods = append(failedPods, unscheduledPods...)
	}
	//result.UnscheduledPods = failedPods
	//sim.ClusterAnalysis(result)

	return &simontype.SimulateResult{
		UnscheduledPods: failedPods,
		NodeStatus:      sim.GetClusterNodeStatus(),
	}, nil
}
