package exportermgr

import (
	"context"
	"fmt"
	"os"
	"io/ioutil"
	"regexp"

	"sigs.k8s.io/yaml"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  appsv1 "k8s.io/api/apps/v1"
  "k8s.io/apimachinery/pkg/labels"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
)

type K8s struct {
	*Config
	clientset kubernetes.Interface
	k8sconfig *rest.Config
}

type K8sLabel struct {
	Label string
	Value string
}

func (k *K8s) SetClient(clientset kubernetes.Interface) {
	k.clientset = clientset
}

func (k *K8s) Client() kubernetes.Interface {
	return k.clientset
}

func (k *K8s) SetK8sConfig(k8sconfig *rest.Config) {
	k.k8sconfig = k8sconfig
}
// Reads K8s config from in cluster only
func (k *K8s) K8sConfig() *rest.Config {
	return k.k8sconfig
}

// Connect
func (k *K8s) Connect() (kubernetes.Interface, error) {
	clientset := k.Client()
	if clientset != nil {
		return clientset, nil
	}
	config := k.K8sConfig()
	if config == nil {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil,err
		}
		if config == nil {
			return nil, fmt.Errorf("'rest.InClusterConfig()' gave us nothing")
		}
		k.SetK8sConfig(config)
	}
	clientset, err := kubernetes.NewForConfig(k.K8sConfig())
	if err != nil {
		return nil, err
	}
	if clientset == nil {
		return nil, fmt.Errorf("'kubernetes.NewForConfig()' gave us nothing")
	}
	k.SetClient(clientset)
	return clientset, nil
}
// Fetch
func (k *K8s) Fetch() (*[]SrvInstance, error) {
	clientset, err := k.Connect()
	if err != nil {
		return nil, err
	}
	if clientset == nil {
		return nil, fmt.Errorf("'Connect()' gave us nothing")
	}
	apps := clientset.AppsV1()
	if apps == nil {
		return nil, fmt.Errorf("'clientset.AppsV1()' didn't return anything")
	}
	deploymentsClient := apps.Deployments(k.K8sNamespace())
	deploymentList, err := deploymentsClient.List(context.TODO(),metav1.ListOptions{
		LabelSelector: labels.FormatLabels(*k.K8sLabels()),
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}
	srvinstances,err := deploymentList2SrvInstances(deploymentList)
	if err != nil {
		return nil,err
	}
	return srvinstances,nil
}
// Remove
func (k *K8s) Remove(deploymentName string) (error) {
	clientset, err := k.Connect()
	if err != nil {
		return err
	}
	deploymentsClient := clientset.AppsV1().Deployments(k.K8sNamespace())
	deletePolicy := metav1.DeletePropagationForeground
	err = deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	return err
}
// Create
func (k *K8s) Create(deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	// Don't create anything without our labels on it
	if labels.AreLabelsInWhiteList(deployment.ObjectMeta.Labels,*k.K8sLabels()) != true {
		return nil, fmt.Errorf("Deployment missing labels: %s",labels.FormatLabels(*k.K8sLabels()))
	} 
	clientset, err := k.Connect()
	if err != nil {
		return nil,err
	}
	deploymentsClient := clientset.AppsV1().Deployments(k.K8sNamespace())
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	return result, err
}
// Read k8s deployment yaml files and turn the into appsv1.Deployment 
func cfg2Object(filename string) (*appsv1.Deployment,error) {
	var deployment *appsv1.Deployment
	_,err := os.Stat(filename)
	if err != nil {
		return deployment, err
	}
	cfg,err := ioutil.ReadFile(filename)
	if err != nil {
		return deployment, err
	}
	err = yaml.Unmarshal(cfg, &deployment)
	return deployment, err
}
// Convert appsv1.Deployment into []SrvInstance
func deploymentList2SrvInstances(deploylist *appsv1.DeploymentList) (*[]SrvInstance,error) {
  var srvinstances []SrvInstance
	for _,d := range deploylist.Items {
		name := ""
		addr := ""
		var err error
		if containers := d.Spec.Template.Spec.Containers; containers != nil {
			if args := containers[0].Args; args != nil {
				if l := len(args); l == 2 {
					addr,err = stripArgs4Addr(args)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		if addr == "" {
			return nil, fmt.Errorf("K8s deployment missing second Arg")
		}
		if name = d.ObjectMeta.Name; name == "" {
			return nil, fmt.Errorf("K8s deployment had no name")
		}
		srvinstances = append(srvinstances,SrvInstance{Name: name, Addr: addr })		
	}
	return &srvinstances,nil
}
// Find Addr in second Arg of K8s deployment
func stripArgs4Addr(args []string ) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("Missing second argument in Args")
	}
	uri := args[1]
	re, err := regexp.Compile("https?://([^:/]+)(?::|/).*")
	if err != nil {
		panic("I can't write a regexp")
	}
	addr := ""
	result := re.FindStringSubmatch(uri)
	if len(result) > 1 {
		addr = result[1]
	}
	if addr == "" {
		return "", fmt.Errorf("K8s Args can't be parsed")
	}
	return addr,nil
}

