/*
Copyright 2019 kubeflow.org.

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

package hdfs

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestHDFS(t *testing.T) {
	scenarios := map[string]struct {
		token              bool
		secret              *v1.Secret
		expectedVolume      v1.Volume
		expectedVolumeMounts []v1.VolumeMount
	}{
		"HDFSToken": {
			token: true,
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "user-hdfs-token",
				},
				Data: map[string][]byte{
					HDFSSecretVolumeName: {},
				},
			},
			expectedVolumeMounts: []v1.VolumeMount{
				{
					Name:      HDFSSecretVolumeName,
					ReadOnly:  true,
					MountPath: HDFSSecretVolumePath,
				},
			},
			expectedVolume: v1.Volume{
				Name: HDFSSecretVolumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						DefaultMode: Int32Ptr(420),
						SecretName: "user-hdfs-token",
					},
				},
			},
		},
		"HDFSConfig": {
			token: false,
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "user-hdfs-cmap",
				},
				Data: map[string][]byte{
					HDFSConfigMapVolumeName: {},
				},
			},
			expectedVolumeMounts: []v1.VolumeMount{
				{
					Name:      HDFSConfigMapVolumeName,
					ReadOnly:  true,
					MountPath: HDFSConfigMapVolumePath,
				},
				{
					Name:      HDFSConfigMapVolumeName,
					SubPath:   "krb5.conf",
					ReadOnly:  true,
					MountPath: HDFSKRBConfPath,
				},
			},
			expectedVolume: v1.Volume{
				Name: HDFSConfigMapVolumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						DefaultMode: Int32Ptr(420),
						SecretName: "user-hdfs-cmap",
					},
				},
			},
		},
	}

	for name, scenario := range scenarios {
		var volume v1.Volume
		var volumeMount v1.VolumeMount
		var volumeMounts []v1.VolumeMount
		if scenario.token {
			volume, volumeMount = BuildTokenVolume(scenario.secret)
			volumeMounts = []v1.VolumeMount{volumeMount}
		} else {
			volume, volumeMounts = BuildCMapVolume(scenario.secret)
		}

		if diff := cmp.Diff(scenario.expectedVolume, volume); diff != "" {
			t.Errorf("Test %q unexpected volume (-want +got): %v", name, diff)
		}

		if diff := cmp.Diff(scenario.expectedVolumeMounts, volumeMounts); diff != "" {
			t.Errorf("Test %q unexpected volumeMount (-want +got): %v", name, diff)
		}
	}
}
