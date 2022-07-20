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

	v1 "github.com/crossplane/crossplane/apis/pkg/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Wait for Provider to be successfully installed.
func WaitForAllProvidersInstalled(ctx context.Context, c client.Client, interval time.Duration, timeout time.Duration, t *testing.T) error {
	if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ProviderList{}
		if err := c.List(ctx, l); err != nil {
			return false, err
		}

		if len(l.Items) != 1 {
			t.Log("The no. of providers installed is not equal to 1")
			for i, item := range l.Items{
				t.Logf("Provider %v : %v", i, item.Name)
			}
			return false, nil
		}

		for _, item := range l.Items {
			if item.GetCondition(v1.TypeInstalled).Status != corev1.ConditionTrue {
				t.Logf("The type of provider %v installed is %v", item.Name, item.TypeMeta)
				return false, nil
			}
			if item.GetCondition(v1.TypeHealthy).Status != corev1.ConditionTrue {
				t.Logf("The status of provider %v installed is %v", item.Name, item.Status)
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

// Wait for Provider to be successfully updated.
func WaitForRevisionTransition(ctx context.Context, c client.Client, p2 string, p1 string, interval time.Duration, timeout time.Duration, t *testing.T) error {
	if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ProviderRevisionList{}
		if err := c.List(ctx, l); err != nil {
			return false, err
		}

		// There should be a revision present for the initial revision and the upgrade.
		if len(l.Items) != 2 {
			t.Log("The no. of provider revisions is not equal to 2")
			for i, item := range l.Items{
				t.Logf("Provider revision %v : %v uses package %v", i, item.Name, item.Spec.Package)
			}
			return false, nil
		}
		for _, item := range l.Items {
			// New ProviderRevision should be Active.
			if item.Spec.Package == p2 && item.GetDesiredState() != v1.PackageRevisionActive {
				t.Logf("The state of new provider revision %v built from package %v and having version %v is %v", item.Name, item.Spec.Package, item.Spec.Revision, item.Status)
				return false, nil
			}
			// Old ProviderRevision should be Inactive.
			if item.Spec.Package == p1 && item.GetDesiredState() != v1.PackageRevisionInactive {
				t.Logf("The state of old provider revision %v built from package %v and having version %v is %v", item.Name, item.Spec.Package, item.Spec.Revision, item.Status)
				return false, nil
			}
			// Both ProviderRevisions should be healthy.
			if item.GetCondition(v1.TypeHealthy).Status != corev1.ConditionTrue {
				t.Logf("The condition of provider revision %v built from package %v and having version %v is %v", item.Name, item.Spec.Package, item.Spec.Revision, item.Status.ConditionedStatus)
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

// Wait for Provider to be successfully deleted.
func WaitForAllProvidersDeleted(ctx context.Context, c client.Client, interval time.Duration, timeout time.Duration, t *testing.T) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ProviderList{}
		if err := c.List(ctx, l); err != nil {
			return false, err
		}
		for _, item := range l.Items {
			t.Log("Undeleted providers :")
			t.Logf("Name: %v \t Type: %v \t Uses Package: %v", item.Name, item.Kind, item.Spec.Package)
		}

		return len(l.Items) == 0, nil
	})
}
