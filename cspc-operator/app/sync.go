package app

import (
	"fmt"
	"github.com/openebs/maya/pkg/version"
	"github.com/pkg/errors"
	apis "github.com/sonasingh46/apis/pkg/apis/cstor/v1"
	cstorintapis "github.com/sonasingh46/apis/pkg/intapis/apis/cstor"
	cstorintapisv1 "github.com/sonasingh46/apis/pkg/intapis/apis/cstor/v1"
	"github.com/sonasingh46/cstor-operator/pkg/cspc/algorithm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
)

const (
	StoragePoolClaimCPK = "openebs.io/storage-pool-claim"
	// CStorPoolClusterCPK is the CStorPoolcluster label
	CStorPoolClusterCPK = "openebs.io/cstor-pool-cluster"
	OpenEBSVersionKey   = "openebs.io/version"
)

func (c *Controller) sync(cspc *apis.CStorPoolCluster, cspiList *apis.CStorPoolInstanceList) error {
	// cleaning up CSPI resources in case of removing poolSpec from CSPC
	// or manual CSPI deletion
	fmt.Println("[DEBUG]: DO NOT FORGET TO REBUILD IF YOUR ARE DEBUUGGING------------------------------->")

	if cspc.DeletionTimestamp.IsZero() {

	}
	cspcObj := cspc
	//cspcObj, err := c.populateVersion(cspcObj)
	//if err != nil {
	//	klog.Errorf("failed to add versionDetails to CSPC %s:%s", cspcObj.Name, err.Error())
	//	return nil
	//}

	// If CSPC is deleted -- handle the deletion.
	if !cspcObj.DeletionTimestamp.IsZero() {

	}

	// Add finalizer on CSPC

	// Convert CSPC external API type to internal type
	// This conversion is done here so that the rest of the code for CSPC related stuff
	// is loosely coupled from the external versioned CSPC type.
	cspcInternal := &cstorintapis.CStorPoolCluster{}
	cstorintapisv1.Convert_v1_CStorPoolCluster_To_cstor_CStorPoolCluster(cspc, cspcInternal, nil)

	// Create pools if required.
	if len(cspiList.Items) < len(cspc.Spec.Pools) {
		return c.ScaleUp(cspcInternal, len(cspcInternal.Spec.Pools)-len(cspiList.Items))
	}

	if len(cspiList.Items) > len(cspc.Spec.Pools) {
		// Scale Down and return
		return c.ScaleDown()
	}

	// Create pool deployment for the CSPIs

	// Handle Pool operations
	return nil
}

// ScaleUp creates as many cstor pool on a node as pendingPoolCount.
func (c *Controller) ScaleUp(cspc *cstorintapis.CStorPoolCluster, pendingPoolCount int) error {
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		err := c.CreateCSPI(cspc)
		if err != nil {
			message := fmt.Sprintf("Pool provisioning failed for %d/%d ", poolCount, pendingPoolCount)
			c.Record(cspc,corev1.EventTypeWarning,"Create",message)
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name))
		} else {
			message := fmt.Sprintf("Pool Provisioned %d/%d ", poolCount, pendingPoolCount)
			c.Record(cspc,corev1.EventTypeNormal,"Create",message)
			klog.Infof("Pool provisioned successfully %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name)
		}
	}
	return nil
}

// CreateCSPI creates CSPI
func (c *Controller) CreateCSPI(cspc *cstorintapis.CStorPoolCluster) error {
	ac, err := algorithm.NewBuilder().
		WithCSPC(cspc).
		WithNameSpace(cspc.Namespace).
		WithKubeClient(c.kubeclientset).
		WithOpenEBSClient(c.clientset).
		Build()
	if err != nil {
		return err
	}
	cspi, err := ac.GetCSPSpec()
	if err != nil {
		return err
	}
	cspiApiObj := &apis.CStorPoolInstance{}
	cstorintapisv1.Convert_cstor_CStorPoolInstance_To_v1_CStorPoolInstance(cspi, cspiApiObj, nil)
	c.clientset.CstorV1().CStorPoolInstances(cspc.Namespace).Create(cspiApiObj)
	return nil
}

// CreateStoragePool creates the required resource to provision a cStor pool
func (c *Controller) CreateCSPIDeployment() error {
	return nil
}

func (c *Controller) ScaleDown() error {
	return nil
}

// populateVersion assigns VersionDetails for old cspc object and newly created
// cspc
func (c *Controller) populateVersion(cspc *apis.CStorPoolCluster) (*apis.CStorPoolCluster, error) {
	if cspc.VersionDetails.Status.Current == "" {
		var err error
		var v string
		var obj *apis.CStorPoolCluster
		v, err = c.EstimateCSPCVersion(cspc)
		if err != nil {
			return nil, err
		}
		cspc.VersionDetails.Status.Current = v
		// For newly created spc Desired field will also be empty.
		cspc.VersionDetails.Desired = v
		cspc.VersionDetails.Status.DependentsUpgraded = true
		obj, err = c.clientset.CstorV1().
			CStorPoolClusters("openebs"). // ToDO: Fix Hardcoding
			Update(cspc)

		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to update spc %s while adding versiondetails",
				cspc.Name,
			)
		}
		klog.Infof("Version %s added on spc %s", v, cspc.Name)
		return obj, nil
	}
	return cspc, nil
}

// EstimateCSPCVersion returns the cspi version if any cspi is present for the cspc or
// returns the maya version as the new cspi created will be of maya version
func (c *Controller) EstimateCSPCVersion(cspc *apis.CStorPoolCluster) (string, error) {
	cspiList, err := c.clientset.CstorV1().
		CStorPoolInstances("openebs"). // ToDO: Fix Hardcoding
		List(
			metav1.ListOptions{
				LabelSelector: string(CStorPoolClusterCPK) + "=" + cspc.Name,
			})
	if err != nil {
		return "", errors.Wrapf(
			err,
			"failed to get the cstorpool instance list related to cspc : %s",
			cspc.Name,
		)
	}
	if len(cspiList.Items) == 0 {
		return version.Current(), nil
	}
	return cspiList.Items[0].Labels[string(OpenEBSVersionKey)], nil
}

func (c *Controller) Record(cspc *cstorintapis.CStorPoolCluster, eventype, reason, message string) {
	cspcAPI := &apis.CStorPoolCluster{}
	cstorintapisv1.Convert_cstor_CStorPoolCluster_To_v1_CStorPoolCluster(cspc, cspcAPI, nil)
	c.recorder.Event(cspcAPI, eventype, reason, message)
}
