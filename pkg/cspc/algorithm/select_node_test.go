/*
Copyright 2018 The OpenEBS Authors

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
	corev1 "k8s.io/api/core/v1"
	core "k8s.io/client-go/testing"
	"strconv"

	openebsFakeClientset "github.com/sonasingh46/apis/pkg/client/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

type fixture struct {
	kubeclient *fake.Clientset
	openebsClient *openebsFakeClientset.Clientset
	kubeObjects []runtime.Object
	openebsObject []runtime.Object
	actions []core.Action
}

func (f *fixture)WithOpenEBSObjects(objects ...runtime.Object)*fixture  {
	f.openebsObject=objects
	f.openebsClient=openebsFakeClientset.NewSimpleClientset(objects...)
	return f
}

func (f *fixture)WithKubeObjects(objects ...runtime.Object)*fixture  {
	f.kubeObjects=objects
	f.kubeclient=fake.NewSimpleClientset(objects...)
	return f
}

func NewNodeList() *corev1.NodeList {
	newNodeList:= &corev1.NodeList{}
	for i:=1;i<4;i++{
		newNode:=&corev1.Node{}
		newNode.Name="node"+strconv.Itoa(i)
		newNode.Labels= map[string]string{"kubernetes.io/hostname":"node"+strconv.Itoa(i)}
		newNodeList.Items=append(newNodeList.Items,*newNode)
	}
	return newNodeList
}

func NewFixture()*fixture  {
	return &fixture{
		kubeclient:fake.NewSimpleClientset(),
		openebsClient:openebsFakeClientset.NewSimpleClientset(),
	}
}


func TestConfig_GetNodeFromLabelSelector(t *testing.T) {
	fixture:=NewFixture().WithKubeObjects(NewNodeList())
	type args struct {
		labels map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:"[node1 exists] Select node with labels : {kubernetes.io/hostname:node1}",
			args:args{
				labels: map[string]string{
					"kubernetes.io/hostname":"node1",
				},
			},
			want:"node1",
			wantErr:false,

		},

		{
			name:"[node2 exists] Select node with labels : {kubernetes.io/hostname:node2}",
			args:args{
				labels: map[string]string{
					"kubernetes.io/hostname":"node2",
				},
			},
			want:"node2",
			wantErr:false,

		},

		{
			name:"[node2 exists] Select node with labels : {kubernetes.io/hostname:node2, dummy.io/dummy:dummy}",
			args:args{
				labels: map[string]string{
					"kubernetes.io/hostname":"node2",
					"dummy.io/dummy":"dummy",
				},
			},
			want:"",
			wantErr:true,

		},

		{
			name:"[node4 does not exist] Select node with labels : {kubernetes.io/hostname:node4}",
			args:args{
				labels: map[string]string{
					"kubernetes.io/hostname":"node4",
				},
			},
			want:"",
			wantErr:true,

		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &Config{
				kubeclientset: fixture.kubeclient,
			}
			got, err := ac.GetNodeFromLabelSelector(tt.args.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.GetNodeFromLabelSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Config.GetNodeFromLabelSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}
