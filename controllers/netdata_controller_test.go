// Copyright 2023 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*


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

package controllers

import (
	"fmt"

	ipamv1alpha1 "github.com/onmetal/ipam/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Netdata Controller delete expired", func() {
	var ns string
	BeforeEach(func(ctx SpecContext) {
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "testns-"},
		}
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
		ns = namespace.Name
		DeferCleanup(func(ctx SpecContext) {
			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
		})
	})

	var subnet *ipamv1alpha1.Subnet
	var res reconcile.Result
	var err error
	BeforeEach(func() {
		subnet = &ipamv1alpha1.Subnet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: ipamv1alpha1.GroupVersion.String(),
				Kind:       "Subnet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      "subnettest",
				Labels: map[string]string{
					"labelsubnet": "oob",
				},
			},
			Spec: ipamv1alpha1.SubnetSpec{
				Network: v1.LocalObjectReference{Name: "test"},
			},
		}
		res = reconcile.Result{}
		err = nil
	})

	When("Subnet has no label labelsubnet", func() {

		JustBeforeEach(func(ctx SpecContext) {
			delete(subnet.Labels, "labelsubnet")
			netdataReconciler.disable()

			// Create a subnet and test if it is created successfully
			Expect(k8sClient.Create(ctx, subnet)).To(Succeed())
			Eventually(func(g Gomega, ctx SpecContext) {
				var obj ipamv1alpha1.Subnet
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: "subnettest"}, &obj)).To(Succeed())
			}, ctx, "3s").Should(Succeed())

			netdataReconciler.enable()
			res, err = netdataReconciler.Reconcile(ctx, controllerruntime.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: "subnettest"}})

		})

		It("Test valid subnet label ", func(ctx SpecContext) {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Not reconciling as Labelsubnet do not match for subnet : subnettest"))
			Expect(res).To(Equal(reconcile.Result{}))
		})
	})

	When("No valid subnet found", func() {
		JustBeforeEach(func(ctx SpecContext) {
			res, err = netdataReconciler.Reconcile(ctx, controllerruntime.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: "subnettest"}})
			fmt.Println(res, err)
		})

		It("Test valid subnet", func(ctx SpecContext) {
			Expect(err).To(MatchError("cannot get Subnet: Subnet.ipam.onmetal.de \"subnettest\" not found"))
			Expect(res).To(Equal(reconcile.Result{}))
		})
	})

})
