package healthcheck

import (
	"context"
	"os"
	"time"

	"github.com/IBM/autopilot/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func ListPVC() (string, error) {
	pvc, err := utils.GetClientsetInstance().Cset.CoreV1().PersistentVolumeClaims(utils.Namespace).Get(context.Background(), utils.PodName, metav1.GetOptions{})
	if err != nil {
		klog.Error("Error in creating the lister", err.Error())
		return "ABORT", err
	}
	switch pvc.Status.Phase {
	case "Bound":
		{
			klog.Info("[PVC Create-Delete] PVC Bound: SUCCESS")
			klog.Info("Observation: ", utils.NodeName, " 0")
			utils.HchecksGauge.WithLabelValues("pvc", utils.NodeName, utils.CPUModel, utils.GPUModel, "").Set(0)
		}
	case "Pending":
		{
			waitonpvc := time.NewTicker(time.Minute)
			defer waitonpvc.Stop()
			<-waitonpvc.C
			pvc, err := utils.GetClientsetInstance().Cset.CoreV1().PersistentVolumeClaims(utils.Namespace).Get(context.Background(), utils.PodName, metav1.GetOptions{})
			if err != nil {
				klog.Error("[PVC Create-Delete] Error in creating the lister: ", err.Error())
				return "[PVC Create-Delete] PVC not found. ABORT ", err
			}
			phase := pvc.Status.Phase
			if pvc.Status.Phase == "Pending" {
				klog.Info("[PVC Create-Delete] Timer is up with PVC Pending. Force delete. FAIL")
				klog.Info("Observation: ", utils.NodeName, " 1")
				utils.HchecksGauge.WithLabelValues("pvc", utils.NodeName, utils.CPUModel, utils.GPUModel, "").Set(1)
				err := deletePVC(utils.PodName)
				if err != nil {
					return "[PVC Create-Delete] Error in deleting the PVC. ABORT ", err
				}
				HealthCheckStatus[PVC] = true
				return "[PVC Create-Delete] FAIL", nil
			}
			if phase == "Bound" {
				klog.Info("[PVC Create-Delete] PVC Bound: SUCCESS")
				klog.Info("Observation: ", utils.NodeName, " 0")
				utils.HchecksGauge.WithLabelValues("pvc", utils.NodeName, utils.CPUModel, utils.GPUModel, "").Set(0)
			}
		}
	}
	err = deletePVC(utils.PodName)
	if err != nil {
		return "Error in deleting the PVC. ABORT ", err
	}
	return "[PVC Create-Delete] PVC SUCCESS", nil
}

func deletePVC(pvc string) error {
	cset := utils.GetClientsetInstance()
	err := cset.Cset.CoreV1().PersistentVolumeClaims(utils.Namespace).Delete(context.TODO(), pvc, metav1.DeleteOptions{})
	if err != nil {
		klog.Info("[PVC Delete] Failed. ABORT. ", err.Error())
	}
	return err
}

func createPVC() error {
	cset := utils.GetClientsetInstance()
	storageclass := os.Getenv("PVC_TEST_STORAGE_CLASS")
	pvcTemplate := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.PodName,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageclass,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": resource.MustParse("100Mi"),
				},
			},
		},
	}
	// Check if any previous instance exists, cleanup if so
	pvc, _ := utils.GetClientsetInstance().Cset.CoreV1().PersistentVolumeClaims(utils.Namespace).Get(context.Background(), utils.PodName, metav1.GetOptions{})

	if pvc.Name != "" {
		klog.Info("[PVC Create] Found pre-existing instance. Cleanup ", pvc.Name)
		deletePVC(utils.PodName)
		waitDelete := time.NewTimer(30 * time.Second)
		<-waitDelete.C
	}

	_, err := cset.Cset.CoreV1().PersistentVolumeClaims(utils.Namespace).Create(context.TODO(), &pvcTemplate, metav1.CreateOptions{})

	if err != nil {
		klog.Info("[PVC Create] Failed. ABORT. ", err.Error())
	}
	return err
}
