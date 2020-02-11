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
	"github.com/sonasingh46/apis/pkg/apis/cstor/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"time"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cspcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncCSPC(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing cstorpoolcluster %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing cstorpoolcluster %q (%v)", key, time.Since(startTime))
	}()

	// Convert the namespace/name string into a distinct namespace and name
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	// Get the cspc resource with this namespace/name
	cspc, err := c.cspcLister.CStorPoolClusters(ns).Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("cspc '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	// TODO: Deep-copy only when needed.
	cspcGot := cspc.DeepCopy()
	cspiList,_:=c.GetCSPIListForCSPC(cspcGot)
	err = c.sync(cspcGot,cspiList)
	return err
}

func (c *Controller) enqueue(cspc *v1.CStorPoolCluster) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(cspc); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller)GetCSPIListForCSPC(cspc *v1.CStorPoolCluster) (*v1.CStorPoolInstanceList,error) {
	return c.clientset.CstorV1().CStorPoolInstances(cspc.Namespace).List(v12.ListOptions{})
}
