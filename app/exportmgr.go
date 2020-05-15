package exportermgr

import (
  "fmt"

  appsv1 "k8s.io/api/apps/v1"
  "github.com/mitchellh/mapstructure"
)

type serviceinterface interface {
  Fetch() (*[]SrvInstance, error)
}

type k8sinterface interface {
	Create(*appsv1.Deployment) (*appsv1.Deployment, error)
	Remove(string) error
	Fetch() (*[]SrvInstance, error)
}

type ExporterMgr struct {
	*Config
	k8s k8sinterface
	ec2client ec2clientinterface
}
type Service struct {
	SrvType string
	Srv interface{}
} 

type SrvInstance struct {
	Name string
	Addr string
}

func (e *ExporterMgr) SetEc2Client(ec2client ec2clientinterface) {
	e.ec2client = ec2client
}
func (e *ExporterMgr) Ec2Client() (ec2clientinterface) {
	return e.ec2client
}
func (e *ExporterMgr) SetK8s(k8s k8sinterface) {
	e.k8s = k8s
}
func (e *ExporterMgr) K8s() (k8sinterface) {
	return e.k8s
}
// Turn config into serviceinterface
func (e *ExporterMgr) mapSrv(s Service) (serviceinterface,error) {
		switch s.SrvType {
		case "Ec2":
			var ec2 *Ec2
			err := mapstructure.Decode(s.Srv,&ec2)
			if err != nil {
				return nil, err 
			}
			if ec2client := e.Ec2Client(); ec2client != nil {
				ec2.SetEc2Client(ec2client)
			}
			return ec2,err
		default:
			return nil,fmt.Errorf("No Service Type Set: %s", s.SrvType)
		}
}
// The Main Show
func (e *ExporterMgr) Run() error {
	// Loop through services defined in config
	for servicename,service := range *e.SerVices() {
		srvinterface,err := e.mapSrv(service)
		if err != nil {
			return err
		}
		// Find things in a service like ec2 that need exporters
		srvintances,err := srvinterface.Fetch()
		if err != nil {
			return err
		}
		// Find exporters already in k8s
		deploysrvinstances,err := e.k8s.Fetch()
		if err != nil {
			return err
		}
		// Join the above lists to figure out what needs to be removed or added
		remove,add,_ := gregorianJoin(*deploysrvinstances,*srvintances)
		if err != nil {
			return err
		}
		// Remove exporters in k8s
		for _,r := range remove {
			err := e.k8s.Remove(r.Name)
			if err != nil {
				return err
			}
		}
		// Add exporters in k8s
		path := e.K8sDeployTemplatesPath()
		cfgfile := fmt.Sprintf("%s/%s.yml",path,servicename)
		deployment,err := cfg2Object(cfgfile)
		if err != nil {
			return err
		}
		for _,a := range add {
			deployment_name := fmt.Sprintf("%s-%s",service,a.Name)
			deployment.ObjectMeta.Name = deployment_name
			deployment_arg := fmt.Sprintf("http://%s:8080/server-status?auto",a.Addr)
			deployment.Spec.Template.Spec.Containers[0].Args[1] = deployment_arg 
			_,err = e.k8s.Create(deployment)
			if err != nil {
				return err
			}	
		}
	}
	return nil
}
// New
func New(c Config) *ExporterMgr {
	new := &ExporterMgr{
		Config: &c,
		k8s: &K8s{Config: &c},
	}
	return new
}

// Given two lists, it tells you what's in a only, b only or both
func gregorianJoin(a []SrvInstance, b []SrvInstance) ([]SrvInstance,[]SrvInstance,[]SrvInstance) {
	j := map[SrvInstance]int{}
	aonly := []SrvInstance{}
	bonly := []SrvInstance{}
	both := []SrvInstance{}
	for _, v := range a { j[v] = 1 }
	for _,v := range b { 
		if _,ok := j[v]; ok { delete(j,v); both = append(both,v) 
		} else { bonly = append(bonly,v) } 
	}
	for k,_ := range j { aonly = append(aonly,k) }
	return aonly,bonly,both
}


