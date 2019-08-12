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
	"fmt"
	"k8s.io/api/core/v1"
)

type HDFSConfig struct {
	HDFSTokenSecret       string `json:"hdfsTokenSecret"`
	HDFSConfigMap 		  string `json:"hdfsConfigMap"`
}

const (
	HDFSLocation                   = "HADOOP_TOKEN_FILE_LOCATION"
	HDFSConfDir                    = "HADOOP_CONF_DIR"
	HDFSUserName                   = "HADOOP_USER_NAME"
	HDFSConfigMapVolumeName        = "hadoop-config-map"
	HDFSSecretVolumeName           = "hadoop-secret"

	HDFSConfigMapVolumePath 	   = "/hadoop/conf"
	HDFSKRBConfPath 	   		   = "/etc/krb5.conf"
	HDFSSecretVolumePath 	   	   = "/hadoop/secrets"
)

func Int32Ptr(i int) *int32 {
	i32 := int32(i)
	return &i32
}

func BuildTokenVolume(secret *v1.Secret) (v1.Volume, v1.VolumeMount) {
	volume := v1.Volume{
		Name: HDFSSecretVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secret.Name,
				DefaultMode: Int32Ptr(420),
			},
		},
	}
	volumeMount := v1.VolumeMount{
		MountPath: HDFSSecretVolumePath,
		Name:      HDFSSecretVolumeName,
		ReadOnly:  true,
	}
	return volume, volumeMount
}

func BuildCMapVolume(secret *v1.Secret) (v1.Volume, []v1.VolumeMount) {
	volume := v1.Volume{
		Name: HDFSConfigMapVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secret.Name,
				DefaultMode: Int32Ptr(420),
			},
		},
	}
	volumeMounts := []v1.VolumeMount{
		{
			MountPath: HDFSConfigMapVolumePath,
			Name:      HDFSConfigMapVolumeName,
			ReadOnly:  true,
		},
		{
			MountPath: HDFSKRBConfPath,
			Name: HDFSConfigMapVolumeName,
			SubPath: "krb5.conf",
			ReadOnly: true,
		},
	}
	return volume, volumeMounts
}

func BuildHDFSEnvs(namespace string) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  HDFSLocation,
			Value: fmt.Sprintf("%s/%s", HDFSSecretVolumePath, HDFSSecretVolumeName),
		},
		{
			Name:  HDFSConfDir,
			Value: HDFSConfigMapVolumePath,
		},
		{
			Name:  HDFSUserName,
			Value: namespace,
		},
	}
}