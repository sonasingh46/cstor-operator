/*
Copyright 2019 The OpenEBS Authors

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

package algorithm

import (
	"github.com/openebs/maya/pkg/version"
	"github.com/pkg/errors"
	cstor "github.com/sonasingh46/apis/pkg/apis/cstor/v1"
	"github.com/sonasingh46/apis/pkg/apis/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	// StoragePoolKindCSPC holds the value of CStorPoolCluster
	StoragePoolKindCSPC = "CStorPoolCluster"
	ApiVersion= "cstor.openebs.io/v1"
)

// GetCSPSpec returns a CSPI spec that should be created and claims all the
// block device present in the CSPI spec
func (ac *Config) GetCSPSpec() (*cstor.CStorPoolInstance, error) {
	poolSpec, nodeName, err := ac.SelectNode()
	if err != nil || nodeName == "" {
		return nil, errors.Wrap(err, "failed to select a node")
	}

	// ToDo: Move following to mutating webhook.
	if poolSpec.PoolConfig.Resources == nil {
		poolSpec.PoolConfig.Resources = ac.CSPC.Spec.DefaultResources
	}
	if poolSpec.PoolConfig.AuxResources == nil {
		poolSpec.PoolConfig.AuxResources = ac.CSPC.Spec.DefaultAuxResources
	}
	if len(poolSpec.PoolConfig.Tolerations) == 0 {
		poolSpec.PoolConfig.Tolerations = ac.CSPC.Spec.Tolerations
	}

	if poolSpec.PoolConfig.PriorityClassName == "" {
		poolSpec.PoolConfig.PriorityClassName = ac.CSPC.Spec.DefaultPriorityClassName
	}

	cspiLabels := ac.buildLabelsForCSPI(nodeName)

	cspiObj := cstor.NewCStorPoolInstance().
		WithName(ac.CSPC.Name + "-" + rand.String(4)).
		WithNamespace(ac.Namespace).
		WithNodeSelectorByReference(poolSpec.NodeSelector).
		WithNodeName(nodeName).
		WithPoolConfig(poolSpec.PoolConfig).
		WithRaidGroups(poolSpec.DataRaidGroups).
		WithCSPCOwnerReference(GetCSPCOwnerReference(ac.CSPC)).
		WithLabelsNew(cspiLabels).
		WithFinalizer(types.CSPCFinalizer).
		WithNewVersion(version.GetVersion()).
		WithDependentsUpgraded()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build CSPI object for node selector {%v}", poolSpec.NodeSelector)
	}

	err = ac.ClaimBDsForNode(GetBDListForNode(*poolSpec))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to claim block devices for node {%s}", nodeName)
	}
	return cspiObj, nil
}

// buildLabelsForCSPI builds labels for CSPI
func (ac *Config) buildLabelsForCSPI(nodeName string) map[string]string {
	labels := make(map[string]string)
	labels[types.HostNameLabelKey] = nodeName
	labels[string(types.CStorPoolClusterLabelKey)] = ac.CSPC.Name
	labels[string(types.OpenEBSVersionLabelKey)] = version.GetVersion()
	labels[string(types.CASTypeLabelKey)] = types.CasTypeCStor
	return labels
}

func GetCSPCOwnerReference(cspc *cstor.CStorPoolCluster) metav1.OwnerReference {
	trueVal := true
	reference := metav1.OwnerReference{
		APIVersion:         ApiVersion,
		Kind:               StoragePoolKindCSPC,
		UID:                cspc.ObjectMeta.UID,
		Name:               cspc.ObjectMeta.Name,
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
	return reference
}
