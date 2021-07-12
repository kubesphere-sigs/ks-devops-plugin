/*
Copyright 2020 The KubeSphere Authors.

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

package config

import (
	authoptions "devops.kubesphere.io/plugin/pkg/apiserver/authentication/options"
	"devops.kubesphere.io/plugin/pkg/client/cache"
	"devops.kubesphere.io/plugin/pkg/client/devops/jenkins"
	"devops.kubesphere.io/plugin/pkg/client/k8s"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigurationName is the default name of configuration
	DefaultConfigurationName = "kubesphere"
	// DefaultConfigurationFileName is the default filename of configuration
	DefaultConfigurationFileName = DefaultConfigurationName + ".yaml"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/kubesphere"
)

// AuthMode is the auth mode of current project
// TODO add a validation method to verify all the supported values
type AuthMode string

var (
	// AuthModeToken let it use the token directly
	AuthModeToken AuthMode = "token"
)

// Config defines everything needed for apiserver to deal with external services
type Config struct {
	JenkinsOptions        *jenkins.Options                   `json:"devops,omitempty" yaml:"devops,omitempty" mapstructure:"devops"`
	KubernetesOptions     *k8s.KubernetesOptions             `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	RedisOptions          *cache.Options                     `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	AuthenticationOptions *authoptions.AuthenticationOptions `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	AuthMode              AuthMode                           `json:"authMode,omitempty" yaml:"authMode,omitempty" mapstructure:"authMode"`
	JWTSecret             string                             `json:"jwtSecret,omitempty" yaml:"jwtSecret,omitempty" mapstructure:"jwtSecret"`
}

// newConfig creates a default non-empty Config
func New() *Config {
	return &Config{
		JenkinsOptions:        jenkins.NewDevopsOptions(),
		KubernetesOptions:     k8s.NewKubernetesOptions(),
		AuthMode:              AuthModeToken,
		AuthenticationOptions: authoptions.NewAuthenticateOptions(),
	}
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*Config, error) {
	viper.SetConfigName(DefaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := New()

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func (conf *Config) ToMap() map[string]bool {
	result := make(map[string]bool, 0)

	if conf == nil {
		return result
	}

	c := reflect.Indirect(reflect.ValueOf(conf))

	for i := 0; i < c.NumField(); i++ {
		name := strings.Split(c.Type().Field(i).Tag.Get("json"), ",")[0]
		if strings.HasPrefix(name, "-") {
			continue
		}

		if c.Field(i).IsNil() {
			result[name] = false
		} else {
			result[name] = true
		}
	}

	return result
}
