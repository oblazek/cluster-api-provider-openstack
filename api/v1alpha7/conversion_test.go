/*
Copyright 2023 The Kubernetes Authors.

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

package v1alpha7

import (
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/pointer"
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	infrav1 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha8"
	testhelpers "sigs.k8s.io/cluster-api-provider-openstack/test/helpers"
)

// Setting this to false to avoid running tests in parallel. Only for use in development.
const parallel = true

func runParallel(f func(t *testing.T)) func(t *testing.T) {
	if parallel {
		return func(t *testing.T) {
			t.Helper()
			t.Parallel()
			f(t)
		}
	}
	return f
}

func TestFuzzyConversion(t *testing.T) {
	// The test already ignores the data annotation added on up-conversion.
	// Also ignore the data annotation added on down-conversion.
	ignoreDataAnnotation := func(hub conversion.Hub) {
		obj := hub.(metav1.Object)
		delete(obj.GetAnnotations(), utilconversion.DataAnnotation)
	}

	fuzzerFuncs := func(_ runtimeserializer.CodecFactory) []interface{} {
		return []interface{}{
			func(spec *infrav1.OpenStackClusterSpec, c fuzz.Continue) {
				c.FuzzNoCustom(spec)

				// The fuzzer only seems to generate Subnets of
				// length 1, but we need to also test length 2.
				// Ensure it is occasionally generated.
				if len(spec.Subnets) == 1 && c.RandBool() {
					subnet := infrav1.SubnetFilter{}
					c.FuzzNoCustom(&subnet)
					spec.Subnets = append(spec.Subnets, subnet)
				}
			},
		}
	}

	t.Run("for OpenStackCluster", runParallel(utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Hub:              &infrav1.OpenStackCluster{},
		Spoke:            &OpenStackCluster{},
		HubAfterMutation: ignoreDataAnnotation,
		FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))

	t.Run("for OpenStackCluster with mutate", runParallel(testhelpers.FuzzMutateTestFunc(testhelpers.FuzzMutateTestFuncInput{
		FuzzTestFuncInput: utilconversion.FuzzTestFuncInput{
			Hub:              &infrav1.OpenStackCluster{},
			Spoke:            &OpenStackCluster{},
			HubAfterMutation: ignoreDataAnnotation,
			FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
		},
		MutateFuzzerFuncs: []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))

	t.Run("for OpenStackClusterTemplate", runParallel(utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Hub:              &infrav1.OpenStackClusterTemplate{},
		Spoke:            &OpenStackClusterTemplate{},
		HubAfterMutation: ignoreDataAnnotation,
		FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))

	t.Run("for OpenStackClusterTemplate with mutate", runParallel(testhelpers.FuzzMutateTestFunc(testhelpers.FuzzMutateTestFuncInput{
		FuzzTestFuncInput: utilconversion.FuzzTestFuncInput{
			Hub:              &infrav1.OpenStackClusterTemplate{},
			Spoke:            &OpenStackClusterTemplate{},
			HubAfterMutation: ignoreDataAnnotation,
			FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
		},
		MutateFuzzerFuncs: []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))

	t.Run("for OpenStackMachine", runParallel(utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Hub:              &infrav1.OpenStackMachine{},
		Spoke:            &OpenStackMachine{},
		HubAfterMutation: ignoreDataAnnotation,
		FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))

	t.Run("for OpenStackMachine with mutate", runParallel(testhelpers.FuzzMutateTestFunc(testhelpers.FuzzMutateTestFuncInput{
		FuzzTestFuncInput: utilconversion.FuzzTestFuncInput{
			Hub:              &infrav1.OpenStackMachine{},
			Spoke:            &OpenStackMachine{},
			HubAfterMutation: ignoreDataAnnotation,
			FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
		},
		MutateFuzzerFuncs: []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))

	t.Run("for OpenStackMachineTemplate", runParallel(utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Hub:              &infrav1.OpenStackMachineTemplate{},
		Spoke:            &OpenStackMachineTemplate{},
		HubAfterMutation: ignoreDataAnnotation,
		FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))

	t.Run("for OpenStackMachineTemplate with mutate", runParallel(testhelpers.FuzzMutateTestFunc(testhelpers.FuzzMutateTestFuncInput{
		FuzzTestFuncInput: utilconversion.FuzzTestFuncInput{
			Hub:              &infrav1.OpenStackMachineTemplate{},
			Spoke:            &OpenStackMachineTemplate{},
			HubAfterMutation: ignoreDataAnnotation,
			FuzzerFuncs:      []fuzzer.FuzzerFuncs{fuzzerFuncs},
		},
		MutateFuzzerFuncs: []fuzzer.FuzzerFuncs{fuzzerFuncs},
	})))
}

func TestMachineConversionControllerSpecFields(t *testing.T) {
	// This tests that we still do field restoration when the controller modifies ProviderID and InstanceID in the spec

	g := gomega.NewWithT(t)
	scheme := runtime.NewScheme()
	g.Expect(AddToScheme(scheme)).To(gomega.Succeed())
	g.Expect(infrav1.AddToScheme(scheme)).To(gomega.Succeed())

	testMachine := func() *OpenStackMachine {
		return &OpenStackMachine{
			Spec: OpenStackMachineSpec{},
		}
	}

	tests := []struct {
		name              string
		modifyUp          func(*infrav1.OpenStackMachine)
		testAfter         func(*OpenStackMachine)
		expectNetworkDiff bool
	}{
		{
			name: "No change",
		},
		{
			name: "Non-ignored change",
			modifyUp: func(up *infrav1.OpenStackMachine) {
				up.Spec.Flavor = "new-flavor"
			},
			testAfter: func(after *OpenStackMachine) {
				g.Expect(after.Spec.Flavor).To(gomega.Equal("new-flavor"))
			},
			expectNetworkDiff: true,
		},
		{
			name: "Set ProviderID",
			modifyUp: func(up *infrav1.OpenStackMachine) {
				up.Spec.ProviderID = pointer.String("new-provider-id")
			},
			testAfter: func(after *OpenStackMachine) {
				g.Expect(after.Spec.ProviderID).To(gomega.Equal(pointer.String("new-provider-id")))
			},
			expectNetworkDiff: false,
		},
		{
			name: "Set InstanceID",
			modifyUp: func(up *infrav1.OpenStackMachine) {
				up.Spec.InstanceID = pointer.String("new-instance-id")
			},
			testAfter: func(after *OpenStackMachine) {
				g.Expect(after.Spec.InstanceID).To(gomega.Equal(pointer.String("new-instance-id")))
			},
			expectNetworkDiff: false,
		},
		{
			name: "Set ProviderID and non-ignored change",
			modifyUp: func(up *infrav1.OpenStackMachine) {
				up.Spec.ProviderID = pointer.String("new-provider-id")
				up.Spec.Flavor = "new-flavor"
			},
			testAfter: func(after *OpenStackMachine) {
				g.Expect(after.Spec.ProviderID).To(gomega.Equal(pointer.String("new-provider-id")))
				g.Expect(after.Spec.Flavor).To(gomega.Equal("new-flavor"))
			},
			expectNetworkDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := testMachine()

			up := infrav1.OpenStackMachine{}
			g.Expect(before.ConvertTo(&up)).To(gomega.Succeed())

			if tt.modifyUp != nil {
				tt.modifyUp(&up)
			}

			after := OpenStackMachine{}
			g.Expect(after.ConvertFrom(&up)).To(gomega.Succeed())

			if tt.testAfter != nil {
				tt.testAfter(&after)
			}
		})
	}
}
