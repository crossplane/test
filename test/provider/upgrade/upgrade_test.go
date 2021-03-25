// +build e2e_provider

/*
Copyright 2021 The Crossplane Authors.

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

package provider

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	v1 "github.com/crossplane/crossplane/apis/pkg/v1"
)

const (
	providerName           = "provider-gcp"
	initialProviderPackage = "crossplane/provider-gcp:v0.16.0"
	upgradeProviderPackage = "crossplane/provider-gcp:master"
)

func TestProviderUpgrade(t *testing.T) {
	cases := map[string]struct {
		reason string
		body   func() error
	}{
		"UpgradeProviderGCPStableToLatest": {
			reason: "Should be able to successfully update provider-gcp from latest stable to latest development build.",
			body: func() error {
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
				provider := &v1.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Name: providerName,
					},
					Spec: v1.ProviderSpec{
						PackageSpec: v1.PackageSpec{
							Package: initialProviderPackage,
						},
					},
				}
				// Create initial Provider.
				if err := a.Apply(ctx, provider); err != nil {
					return err
				}

				// Wait for Provider to be successfully installed.
				if err := wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
					l := &v1.ProviderList{}
					if err := c.List(ctx, l); err != nil {
						return false, err
					}
					if len(l.Items) != 1 {
						return false, nil
					}
					for _, p := range l.Items {
						if p.GetCondition(v1.TypeInstalled).Status != corev1.ConditionTrue {
							return false, nil
						}
						if p.GetCondition(v1.TypeHealthy).Status != corev1.ConditionTrue {
							return false, nil
						}
					}
					return true, nil
				}); err != nil {
					return err
				}

				// Update Provider package.
				provider.Spec.Package = upgradeProviderPackage
				if err := a.Apply(ctx, provider); err != nil {
					return err
				}

				// Wait for Provider to be successfully updated.
				if err := wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
					l := &v1.ProviderRevisionList{}
					if err := c.List(ctx, l); err != nil {
						return false, err
					}
					// There should be a revision present for the initial revision and the upgrade.
					if len(l.Items) != 2 {
						return false, nil
					}
					for _, p := range l.Items {
						// New ProviderRevision should be Active.
						if p.Spec.Package == upgradeProviderPackage && p.GetDesiredState() != v1.PackageRevisionActive {
							return false, nil
						}
						// Old ProviderRevision should be Inactive.
						if p.Spec.Package == initialProviderPackage && p.GetDesiredState() != v1.PackageRevisionInactive {
							return false, nil
						}
						// Both ProviderRevisions should be healthy.
						if p.GetCondition(v1.TypeHealthy).Status != corev1.ConditionTrue {
							return false, nil
						}
					}
					return true, nil
				}); err != nil {
					return err
				}

				// Clean up Provider.
				if err := c.DeleteAllOf(ctx, provider); err != nil {
					return err
				}

				// Wait for Provider to be successfully deleted.
				return wait.PollImmediate(5*time.Second, 30*time.Second, func() (bool, error) {
					l := &v1.ProviderList{}
					if err := c.List(ctx, l); err != nil {
						return false, err
					}
					return len(l.Items) == 0, nil
				})
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if err := tc.body(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
