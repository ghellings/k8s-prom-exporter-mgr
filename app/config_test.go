package exportermgr

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-test/deep"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func init() { log.SetLevel(log.ErrorLevel) }

func TestConfigSetConfigFile(t *testing.T) {
	config := Config{}
	config.SetConfigFile("testConfigFile")
	if config.ConfigFile() != "testConfigFile" {
		t.Error("Expected 'ConfigFile' to equal 'testConfigFile' and it didn't")
	}
}
func TestConfigSetK8sDeployTemplatesPath(t *testing.T) {
	config := Config{}
	config.SetK8sDeployTemplatesPath("/test/path")
	if config.K8sDeployTemplatesPath() != "/test/path" {
		t.Error("Expect 'K8sDeployTemplatesPath' to equal '/test/path' and it didn't")
	}
}
func TestConfigSetK8sNamespace(t *testing.T) {
	config := Config{}
	config.SetK8sNamespace("testNamespace")
	if config.K8sNamespace() != "testNamespace" {
		t.Error("Expected 'K8sNamespace' to equal 'testNamespace' and it didn't")
	}
}
func TestConfigSetK8sLabels(t *testing.T) {
	config := Config{}
	labels := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}
	config.SetK8sLabels(labels)
	if diff := deep.Equal(config.K8sLabels(), labels); diff != nil {
		t.Errorf("Expected 'K8sLabels' to match what we gave it and it didn't: %s", diff)
	}
}
func TestConfigReadConfig(t *testing.T) {
	config := Config{}
	// Test missing config
	_, err := ReadConfig("bogus")
	if err == nil {
		t.Error("Expect missing file error and didn't get it")
	}
	// Test broken config
	_, err = ReadConfig("config.go")
	if err == nil {
		t.Error("Expected yaml error and didn't get it")
	}
	// Mock a config file for testing
	testpath, err := ioutil.TempDir("", "exportmgr")
	if err != nil {
		t.Error(err)
	}
	testcfgfile := testpath + "/exportmgr.yml"
	defer os.RemoveAll(testpath)
	k8slabels := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}
	services := map[string]interface{}{
		"Apache": &Service{
			SrvType: "Ec2",
			Srv: &Ec2{
				Tags: &[]Ec2Tag{
					{
						Tag:   "tag1",
						Value: "value1",
					},
					{
						Tag:   "tag2",
						Value: "value2",
					},
				},
			},
		},
	}
	testcfg := &Config{
		K8snamespace: "default",
		K8slabels:    k8slabels,
		Services:     services,
	}
	yaml, err := yaml.Marshal(testcfg)
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile(testcfgfile, yaml, 0644)
	if err != nil {
		t.Error(err)
	}
	config, err = ReadConfig(testcfgfile)
	if err != nil {
		t.Errorf("Failed to read config file:%s\n", err)
	}
	if config.ConfigFile() != testcfgfile {
		t.Errorf("Configfile setting expected to be '%s' got '%s'\n", testcfgfile, config.ConfigFile())
	}
	if v := config.K8sNamespace(); v != "default" {
		t.Errorf("Expected K8snamespace to equal 'default' got '%s'", v)
	}
	if diff := deep.Equal(config.K8sLabels(), k8slabels); diff != nil {
		t.Errorf("Expect K8sLabels to match what we gave it : %s", diff)
	}
}
