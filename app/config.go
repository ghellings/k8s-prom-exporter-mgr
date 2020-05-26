package exportermgr

import (
	"io/ioutil"
	"os"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Configfile             string
	K8sdeploytemplatespath string
	K8snamespace           string
	Services               map[string]interface{}
	K8slabels              map[string]string
}

func (c *Config) SetConfigFile(configfile string) {
	c.Configfile = configfile
}
func (c *Config) ConfigFile() string {
	return c.Configfile
}
func (c *Config) SetK8sDeployTemplatesPath(path string) {
	c.K8sdeploytemplatespath = path
}
func (c *Config) K8sDeployTemplatesPath() string {
	return c.K8sdeploytemplatespath
}
func (c *Config) SetK8sNamespace(namespace string) {
	c.K8snamespace = namespace
}
func (c *Config) K8sNamespace() string {
	return c.K8snamespace
}
func (c *Config) SetK8sLabels(labels map[string]string) {
	c.K8slabels = labels
}
func (c *Config) K8sLabels() map[string]string {
	return c.K8slabels
}
func (c *Config) SetSerVices(services map[string]interface{}) {
	c.Services = services
}
func (c *Config) SerVices() map[string]Service {
	var r map[string]Service
	err := mapstructure.Decode(c.Services, &r)
	if err != nil {
		panic(err)
	}
	return r
}
func ReadConfig(filename string) (Config, error) {
	config := Config{}
	config.SetConfigFile(filename)
	_, err := os.Stat(filename)
	if err != nil {
		return config, err

	}
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(cfg, &config)
	config.SetConfigFile(filename)
	return config, err
}
