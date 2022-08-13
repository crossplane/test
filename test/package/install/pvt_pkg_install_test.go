// +build e2e_package

/*
Copyright 2022 The Crossplane Authors.

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

package private_package

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	v1 "github.com/crossplane/crossplane/apis/pkg/v1"
	"github.com/google/go-containerregistry/pkg/name"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pp "github.com/crossplane/test/apis/package"
)

func TestPackageInstall(t *testing.T) {
	cases := map[string]struct {
		reason string
		body   func(pp.PackageConformance) error
	}{
		"InstallPrivatePackage": {
			reason: "Should be able to successfully install a private package from registries supported by Crossplane.",
			body: func(privatePackage pp.PackageConformance) error {
				installedPackageName, err := name.ParseReference(privatePackage.PackageName)
				if err != nil {
					return err
				}

				ctx := context.Background()
				s := runtime.NewScheme()
				if err := v1.AddToScheme(s); err != nil {
					return err
				}

				c, err := client.New(ctrl.GetConfigOrDie(), client.Options{
					Scheme: s,
				})
				if err != nil {
					return err
				}
				a := resource.NewAPIUpdatingApplicator(c)
				p := &v1.Configuration{
					ObjectMeta: metav1.ObjectMeta{
						Name: strings.ReplaceAll(installedPackageName.Context().RegistryStr(), "/", "-"),
					},
					Spec: v1.ConfigurationSpec{
						PackageSpec: v1.PackageSpec{
							Package:            installedPackageName.String(),
							PackagePullSecrets: []corev1.LocalObjectReference{
								corev1.LocalObjectReference{Name: "package-pull-secret"},
							},
						},
					},
				}
				// Create initial Package
				if err := a.Apply(ctx, p); err != nil {
					return err
				}		

				// Wait for Package to be successfully installed.
				if err := CheckIfPackageInstalled(ctx, c, 5*time.Second, 30*time.Second, t); err != nil {
					return err
				}

				// Clean up Provider.
				if err := c.DeleteAllOf(ctx, p); err != nil {
					return err
				}

				// Wait for Packages to be successfully deleted.
				return WaitForAllPackagesDeleted(ctx, c, 5*time.Second, 30*time.Second, t)
			},
		},
	}
	config := pp.GetConfiguration("../../../config/package/conformance.yml")

	for _, pr := range config.Packages {
			for name, tc := range cases {
				t.Run(name, func(t *testing.T) {
					if err := tc.body(pr); err != nil {
						t.Fatal(err)
					}
				})
			}
	}
}

func CheckIfPackageInstalled(ctx context.Context, c client.Client, interval time.Duration, timeout time.Duration, t *testing.T) error {

	if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ConfigurationList{}
		if err := c.List(ctx, l); err != nil {
			return false, err
		}
		for _, item := range l.Items {
			t.Logf("Name: %v \t Type: %v from Registry: %v", item.Name, item.Kind, item.ObjectMeta.Name)
		}

		if len(l.Items) != 1 {
			t.Log("The no. of packages installed is not equal to 1")
			for i, item := range l.Items{
				t.Logf("Package %v : %v", i, item.Name)
			}
			return false, nil
		}
		for _, item := range l.Items {
			if item.GetCondition(v1.TypeInstalled).Status != corev1.ConditionTrue {
				t.Logf("The type of installed package: %v is %v", item.Name, item.TypeMeta)
				return false, nil
			}
			if item.GetCondition(v1.TypeHealthy).Status != corev1.ConditionTrue {
				t.Logf("The status of installed package: %v is %v", item.Name, item.Status)
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

func WaitForAllPackagesDeleted(ctx context.Context, c client.Client, interval time.Duration, timeout time.Duration, t *testing.T) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ConfigurationList{}
		if err := c.List(ctx, l); err != nil {
			return false, err
		}
		for _, item := range l.Items {
			t.Log("Undeleted packages :")
			t.Logf("Name: %v \t Type: %v from Registry: %v", item.Name, item.Kind, item.ObjectMeta.Name)
		}
		return len(l.Items) == 0, nil
	})
}
