package exportermgr

import (
	"testing"
	"os"
	"io/ioutil"

	"github.com/go-test/deep"
	"sigs.k8s.io/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  apiv1 "k8s.io/api/core/v1"
  appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestK8sSetClient(t *testing.T) {
	k8s := K8s{
		Config: &Config{
			K8snamespace: "default",
		},
	}
	client := &kubernetes.Clientset{}
	k8s.SetClient(client)
	checkclient := k8s.Client()
	if diff := deep.Equal(checkclient, client); diff != nil {
		t.Errorf("Expected K8s client and didn't get it")
	}
}

func TestK8sSetK8sConfig(t *testing.T) {
	k8s := K8s{
		Config: &Config{
			K8snamespace: "default",
		},
	}
	config := &rest.Config{}
	k8s.SetK8sConfig(config)
	checkconfig := k8s.K8sConfig()
	if diff := deep.Equal(checkconfig, config); diff != nil {
		t.Errorf("Expected K8s config and didn't get it")
	}
}

func TestK8sConnect(t *testing.T) {
	k8s := K8s{
		Config: &Config{
			K8snamespace: "default",
		},
	}
	fakeclient := fake.NewSimpleClientset()
	k8s.SetClient(fakeclient)
	_,err := k8s.Connect()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestK8sFetch(t *testing.T) {
		deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prom-apache-exporter",
			Namespace: "default",
			Labels: map[string]string{
				"app": "prom-apache-exporter",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "prom-apache-exporter",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "prom-apache-exporter",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "prom-apache-exporter",
							Image: "prom-apache-exporter",
							Command: []string{
								"/apache_exporter/apache_exporter",
							},
							Args: []string{
								"-scrape_uri",
          			"http://10.0.2.85:8080/server-status?auto",
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	fakeclient := fake.NewSimpleClientset(deployment)
	k8s := K8s{
		Config: &Config{
			K8snamespace: "default",
			K8slabels: &map[string]string{
				"app": "prom-apache-exporter",
			},
		},
		clientset: fakeclient,
	}
	srvinstances, err := k8s.Fetch()
	if err != nil {
		t.Errorf(err.Error())
	}
	if length := len(*srvinstances); length != 1 {
		t.Errorf("Expected one deployment got : %d\n", length)
		return
	}
	if (*srvinstances)[0].Name != "prom-apache-exporter" {
		t.Errorf("Expected deployment named 'prom-apache-exporter' got : %s", (*srvinstances)[0].Name)
	}
	// Test that we get nothing if the label doesn't match
	fakeclient = fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prom-apache-exporter",
				Namespace: "default",
				Labels: map[string]string{
					"app": "not-prom-apache-exporter",
				},
			},
		},
	)
	k8s.SetClient(fakeclient)
	srvinstances, err = k8s.Fetch()
	if err != nil {
		t.Errorf(err.Error())
	}
	if length := len(*srvinstances); length != 0 {
		t.Errorf("Expected zero deployment got : %d\n", length)
		return
	}
}

func TestK8sRemove(t *testing.T) {
	k8s := K8s{
		Config: &Config{
			K8snamespace: "default",
		},
	}
	fakeclient := fake.NewSimpleClientset(
		&appsv1.Deployment{
    	ObjectMeta: metav1.ObjectMeta{
        Name:        "prom-apache-exporter",
        Namespace:   "default",
        Labels: map[string]string{
					"app": "prom-apache-exporter",
				},
			},
		},
	)
	k8s.SetClient(fakeclient)
	err := k8s.Remove("prom-apache-exporter")
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestK8sCreate(t *testing.T) {
	k8s := K8s{
		Config: &Config{
			K8snamespace: "default",
			K8slabels: &map[string]string{
				"app": "prom-apache-exporter",
			},
		},
	}
	fakeclient := fake.NewSimpleClientset()
	k8s.SetClient(fakeclient)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prom-apache-exporter",
			Namespace: "default",
			Labels: map[string]string{
				"app": "prom-apache-exporter",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "prom-apache-exporter",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "prom-apache-exporter",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "prom-apache-exporter",
							Image: "prom-apache-exporter",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	checkdeployment, err := k8s.Create(deployment)
	if err != nil {
		t.Errorf(err.Error())
	}
	if diff := deep.Equal(checkdeployment, deployment); diff != nil {
		t.Errorf("Expected to get 'deplyment' back and didn't: %s", diff )
	}
}

func TestK8scfg2Object(t *testing.T) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prom-apache-exporter",
			Namespace: "default",
			Labels: map[string]string{
				"app": "prom-apache-exporter",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "prom-apache-exporter",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "prom-apache-exporter",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "prom-apache-exporter",
							Image: "prom-apache-exporter",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	// Mock a config file for testing
	testpath, err := ioutil.TempDir("","exportmgr")
	if err != nil {
		t.Error(err)
	}
	testcfgfile := testpath+"/prom-apache-exporter.yml"
	defer os.RemoveAll(testpath)
	yaml,err := yaml.Marshal(deployment)
	if err != nil {
		t.Error(err)
	}	
	err = ioutil.WriteFile(testcfgfile, yaml, 0644)
	if err != nil {
		t.Error(err)
	}
	cfgobj, err := cfg2Object(testcfgfile)
	if err != nil {
		t.Error(err)
	}
	if cfgobj.ObjectMeta.Name != deployment.ObjectMeta.Name {
		t.Errorf("Expected Cfg2Object to return object the same as 'deployment' and it didn't: %#v", cfgobj)
	} 

}

func TestK8sdeploymentList2SrvInstances(t *testing.T) {
	deploylist := &appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{Name:"Test"},
				Spec: appsv1.DeploymentSpec{
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								{
									Args: []string{
										"-scrape_uri",
										"http://192.168.1.1:8080/server-status?auto",
									},
								},
							},
						},
					},
				}, 
			},
		},
	}
	srvinstances,err := deploymentList2SrvInstances(deploylist)
	if err != nil {
		t.Error(err)
	}
	testsrvinstances := &[]SrvInstance{
		{
			Name: "Test",
			Addr: "192.168.1.1",
		},
	}
	if diff := deep.Equal(srvinstances,testsrvinstances); diff != nil {
		t.Errorf("Expected to get a &[]SrvInstance and didn't: %#v", srvinstances)
	}
}

func TestK8sstripArgs4Addr(t *testing.T) {
	testargs := []string{
		"-scrape_uri",
		"http://192.168.1.1:8080/server-status?auto",
	}
	addr,err := stripArgs4Addr(testargs)
	if err != nil {
		t.Error(err)
	}
	if addr != "192.168.1.1" {
		t.Errorf("Expected to get '192.168.1.1' and got: %#v", addr)
	}
}

func int32Ptr(i int32) *int32 { return &i }
