/*
Copyright 2020 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"fmt"
	"github.com/sonasingh46/cstor-operator/pkg/cspc/algorithm"
	"k8s.io/klog"
	"github.com/sonasingh46/apis/pkg/apis/types"
	cstor "github.com/sonasingh46/apis/pkg/apis/cstor/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/pkg/errors"

)

// ScaleUp creates as many cstor pool on a node as pendingPoolCount.
func (c *Controller) ScaleUp(cspc *cstor.CStorPoolCluster, pendingPoolCount int) error {
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		err := c.CreateCSPI(cspc)
		if err != nil {
			message := fmt.Sprintf("Pool provisioning failed for %d/%d ", poolCount, pendingPoolCount)
			c.recorder.Event(cspc, corev1.EventTypeWarning, "Create", message)
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name))
		} else {
			message := fmt.Sprintf("Pool Provisioned %d/%d ", poolCount, pendingPoolCount)
			c.recorder.Event(cspc, corev1.EventTypeNormal, "Create", message)
			klog.Infof("Pool provisioned successfully %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name)
		}
	}
	return nil
}

// CreateCSPI creates CSPI
func (c *Controller) CreateCSPI(cspc *cstor.CStorPoolCluster) error {
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
	_, err = c.GetStoredCStorVersionClient().CStorPoolInstances(cspc.Namespace).Create(cspi)

	if err != nil {
		return err
	}
	err=c.CreateCSPIDeployment(cspc,cspi)
	if err!=nil{
		return err
	}
	return nil
}

func (c *Controller) createDeployForCSPList(cspc *cstor.CStorPoolCluster, cspList []cstor.CStorPoolInstance) {
	for _, cspObj := range cspList {
		cspObj := cspObj
		err := c.CreateCSPIDeployment(cspc,&cspObj)
		if err != nil {
			message := fmt.Sprintf("Failed to create pool deployment for CSP %s: %s", cspObj.Name, err.Error())
			c.recorder.Event(cspc, corev1.EventTypeWarning, "PoolDeploymentCreate", message)
			runtime.HandleError(errors.Errorf("Failed to create pool deployment for CSP %s: %s", cspObj.Name, err.Error()))
		}
	}
}
// CreateStoragePool creates the required resource to provision a cStor pool
func (c *Controller) CreateCSPIDeployment(cspc *cstor.CStorPoolCluster,cspi *cstor.CStorPoolInstance) error {
	ac, err := algorithm.NewBuilder().
		WithCSPC(cspc).
		WithNameSpace(cspc.Namespace).
		WithKubeClient(c.kubeclientset).
		WithOpenEBSClient(c.clientset).
		Build()
	if err != nil {
		return err
	}
	deploy,err:=ac.GetPoolDeploySpec(cspi)
	if err != nil {
		return err
	}
	_,err=c.kubeclientset.AppsV1().Deployments(cspi.Namespace).Create(deploy)
	if err!=nil{
		return err
	}
	return nil
}

// DownScalePool deletes the required pool.
func (c *Controller) ScaleDown(cspc *cstor.CStorPoolCluster) error {
	orphanedCSP, err := c.getOrphanedCStorPools(cspc)
	if err != nil {
		c.recorder.Event(cspc, corev1.EventTypeWarning,
			"DownScale", "Pool downscale failed "+err.Error())
		return errors.Wrap(err, "could not get orphaned CSP(s)")
	}
	for _, cspiName := range orphanedCSP {
		c.recorder.Event(cspc, corev1.EventTypeNormal,
			"DownScale", "De-provisioning pool "+cspiName)

		// TODO : As part of deleting a CSP, do we need to delete associated BDCs ?

		err := c.GetStoredCStorVersionClient().CStorPoolInstances(cspc.Namespace).Delete(cspiName, &metav1.DeleteOptions{})
		if err != nil {
			c.recorder.Event(cspc, corev1.EventTypeWarning,
				"DownScale", "De-provisioning pool "+cspiName+"failed")
			klog.Errorf("De-provisioning pool %s failed: %s", cspiName, err)
		}
	}
	return nil
}

// getOrphanedCStorPools returns a list of CSPI names that should be deleted.
func (c *Controller) getOrphanedCStorPools(cspc *cstor.CStorPoolCluster) ([]string, error) {
	var orphanedCSP []string
	nodePresentOnCSPC, err := c.getNodePresentOnCSPC(cspc)
	if err != nil {
		return []string{}, errors.Wrap(err, "could not get node names of pool config present on CSPC")
	}
	cspList, err := c.GetStoredCStorVersionClient().CStorPoolInstances(cspc.Namespace).List(
		metav1.ListOptions{LabelSelector: string(types.CStorPoolClusterLabelKey) + "=" + cspc.Name})

	if err != nil {
		return []string{}, errors.Wrap(err, "could not list CSP(s)")
	}

	for _, cspObj := range cspList.Items {
		cspObj := cspObj
		if nodePresentOnCSPC[cspObj.Spec.HostName] {
			continue
		}
		orphanedCSP = append(orphanedCSP, cspObj.Name)
	}
	return orphanedCSP, nil
}

// getNodePresentOnCSPC returns a map of node names where pool should
// be present.
func (c *Controller) getNodePresentOnCSPC(cspc *cstor.CStorPoolCluster) (map[string]bool, error) {
	nodeMap := make(map[string]bool)
	ac, err := algorithm.NewBuilder().
		WithCSPC(cspc).
		WithNameSpace(cspc.Namespace).
		WithKubeClient(c.kubeclientset).
		WithOpenEBSClient(c.clientset).
		Build()
	if err != nil {
		return nil, err
	}
	for _, pool := range cspc.Spec.Pools {
		nodeName, err := ac.GetNodeFromLabelSelector(pool.NodeSelector)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not get node name for node selector {%v} "+
					"from cspc %s", pool.NodeSelector, cspc.Name)
		}
		nodeMap[nodeName] = true
	}
	return nodeMap, nil
}

