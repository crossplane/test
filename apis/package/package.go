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
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

type Config struct {
	Packages []PackageConformance
}
type PackageConformance struct {
	RegistryName string
	PackageName string
}

func GetConfiguration(path string) Config {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var config Config

	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		panic(err)
	}

	return config
}

