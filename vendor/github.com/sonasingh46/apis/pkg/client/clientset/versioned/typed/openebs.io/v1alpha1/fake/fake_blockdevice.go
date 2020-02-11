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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/sonasingh46/apis/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBlockDevices implements BlockDeviceInterface
type FakeBlockDevices struct {
	Fake *FakeOpenebsV1alpha1
	ns   string
}

var blockdevicesResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "blockdevices"}

var blockdevicesKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "BlockDevice"}

// Get takes name of the blockDevice, and returns the corresponding blockDevice object, and an error if there is any.
func (c *FakeBlockDevices) Get(name string, options v1.GetOptions) (result *v1alpha1.BlockDevice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(blockdevicesResource, c.ns, name), &v1alpha1.BlockDevice{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BlockDevice), err
}

// List takes label and field selectors, and returns the list of BlockDevices that match those selectors.
func (c *FakeBlockDevices) List(opts v1.ListOptions) (result *v1alpha1.BlockDeviceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(blockdevicesResource, blockdevicesKind, c.ns, opts), &v1alpha1.BlockDeviceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BlockDeviceList{ListMeta: obj.(*v1alpha1.BlockDeviceList).ListMeta}
	for _, item := range obj.(*v1alpha1.BlockDeviceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested blockDevices.
func (c *FakeBlockDevices) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(blockdevicesResource, c.ns, opts))

}

// Create takes the representation of a blockDevice and creates it.  Returns the server's representation of the blockDevice, and an error, if there is any.
func (c *FakeBlockDevices) Create(blockDevice *v1alpha1.BlockDevice) (result *v1alpha1.BlockDevice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(blockdevicesResource, c.ns, blockDevice), &v1alpha1.BlockDevice{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BlockDevice), err
}

// Update takes the representation of a blockDevice and updates it. Returns the server's representation of the blockDevice, and an error, if there is any.
func (c *FakeBlockDevices) Update(blockDevice *v1alpha1.BlockDevice) (result *v1alpha1.BlockDevice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(blockdevicesResource, c.ns, blockDevice), &v1alpha1.BlockDevice{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BlockDevice), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeBlockDevices) UpdateStatus(blockDevice *v1alpha1.BlockDevice) (*v1alpha1.BlockDevice, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(blockdevicesResource, "status", c.ns, blockDevice), &v1alpha1.BlockDevice{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BlockDevice), err
}

// Delete takes name of the blockDevice and deletes it. Returns an error if one occurs.
func (c *FakeBlockDevices) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(blockdevicesResource, c.ns, name), &v1alpha1.BlockDevice{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBlockDevices) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(blockdevicesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BlockDeviceList{})
	return err
}

// Patch applies the patch and returns the patched blockDevice.
func (c *FakeBlockDevices) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BlockDevice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(blockdevicesResource, c.ns, name, pt, data, subresources...), &v1alpha1.BlockDevice{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BlockDevice), err
}
