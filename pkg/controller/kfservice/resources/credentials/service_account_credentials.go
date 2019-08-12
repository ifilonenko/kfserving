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

package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"kfserving/pkg/controller/kfservice/resources/credentials/hdfs"

	"github.com/kubeflow/kfserving/pkg/controller/kfservice/resources/credentials/gcs"
	"github.com/kubeflow/kfserving/pkg/controller/kfservice/resources/credentials/s3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	CredentialConfigKeyName = "credentials"
)

type CredentialConfig struct {
	S3  s3.S3Config   `json:"s3,omitempty"`
	GCS gcs.GCSConfig `json:"gcs,omitempty"`
	HDFS hdfs.HDFSConfig `json:"hdfs,omitempty"`
}

type CredentialBuilder struct {
	client client.Client
	config CredentialConfig
}

type HasHadoop struct {
	token bool
	config bool
}

var log = logf.Log.WithName("CredentialBulder")

func NewCredentialBulder(client client.Client, config *v1.ConfigMap) *CredentialBuilder {
	credentialConfig := CredentialConfig{}
	if credential, ok := config.Data[CredentialConfigKeyName]; ok {
		err := json.Unmarshal([]byte(credential), &credentialConfig)
		if err != nil {
			panic(fmt.Errorf("Unable to unmarshall json string due to %v ", err))
		}
	}
	return &CredentialBuilder{
		client: client,
		config: credentialConfig,
	}
}

func (c *CredentialBuilder) CreateSecretVolumeAndEnv(namespace string, serviceAccountName string,
	container *v1.Container, volumes *[]v1.Volume) error {
	if serviceAccountName == "" {
		serviceAccountName = "default"
	}
	s3SecretAccessKeyName := s3.AWSSecretAccessKeyName
	gcsCredentialFileName := gcs.GCSCredentialFileName
	hdfsTokenSecretName := hdfs.HDFSSecretVolumeName
	hdfsConfigMapName := hdfs.HDFSConfigMapVolumeName

	hdfsBool := &HasHadoop{ false, false}

	if c.config.S3.S3SecretAccessKeyName != "" {
		s3SecretAccessKeyName = c.config.S3.S3SecretAccessKeyName
	}

	if c.config.GCS.GCSCredentialFileName != "" {
		gcsCredentialFileName = c.config.GCS.GCSCredentialFileName
	}

	if c.config.HDFS.HDFSTokenSecret != "" {
		hdfsTokenSecretName = c.config.HDFS.HDFSTokenSecret
	}

	if c.config.HDFS.HDFSConfigMap != "" {
		hdfsConfigMapName = c.config.HDFS.HDFSConfigMap
	}

	serviceAccount := &v1.ServiceAccount{}
	err := c.client.Get(context.TODO(), types.NamespacedName{Name: serviceAccountName,
		Namespace: namespace}, serviceAccount)
	if err != nil {
		log.Error(err, "Failed to find service account", "ServiceAccountName", serviceAccountName)
		return nil
	}
	for _, secretRef := range serviceAccount.Secrets {
		secret := &v1.Secret{}
		err := c.client.Get(context.TODO(), types.NamespacedName{Name: secretRef.Name,
			Namespace: namespace}, secret)
		if err != nil {
			log.Error(err, "Failed to find secret", "SecretName", secretRef.Name)
			continue
		}
		if _, ok := secret.Data[s3SecretAccessKeyName]; ok {
			log.Info("Setting secret envs for s3", "S3Secret", secret.Name)
			envs := s3.BuildSecretEnvs(secret, &c.config.S3)
			container.Env = append(container.Env, envs...)
		} else if _, ok := secret.Data[gcsCredentialFileName]; ok {
			log.Info("Setting secret volume for gcs", "GCSSecret", secret.Name)
			volume, volumeMount := gcs.BuildSecretVolume(secret)
			*volumes = append(*volumes, volume)
			container.VolumeMounts =
				append(container.VolumeMounts, volumeMount)
			container.Env = append(container.Env,
				v1.EnvVar{
					Name:  gcs.GCSCredentialEnvKey,
					Value: gcs.GCSCredentialVolumeMountPath + gcsCredentialFileName,
				})
		} else if _, ok := secret.Data[hdfsTokenSecretName]; ok{
			hdfsBool.token = true
			log.Error(fmt.Errorf("you suck"),"hi")
			log.Info("Setting secret volume for hdfs token", "HDFSSecret", secret.Name)
			volume, volumeMount := hdfs.BuildTokenVolume(secret)
			*volumes = append(*volumes, volume)
			container.VolumeMounts =
				append(container.VolumeMounts, volumeMount)
		} else if _, ok := secret.Data[hdfsConfigMapName]; ok {
			hdfsBool.config = true
			log.Error(fmt.Errorf("you suck"),"hi")
			log.Info("Setting secret volume for hdfs cm", "HDFSConfigMap", secret.Name)
			volume, volumeMounts := hdfs.BuildCMapVolume(secret)
			*volumes = append(*volumes, volume)
			container.VolumeMounts =
				append(container.VolumeMounts, volumeMounts...)
		} else {
			log.V(5).Info("Skipping non gcs/s3 secret", "Secret", secret.Name)
		}
	}
	if hdfsBool.token && hdfsBool.config {
		log.Error(fmt.Errorf("you suck"),"hi")
		container.Env = append(container.Env,
			hdfs.BuildHDFSEnvs(namespace)...)
	} else if hdfsBool.token || hdfsBool.config {
		panic(fmt.Errorf("need to define both HDFS Token and ConfigMap"))
	}
	return nil
}
