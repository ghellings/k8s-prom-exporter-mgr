package exportermgr

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-test/deep"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
)

func init() { log.SetLevel(log.ErrorLevel) }

func TestExporterMgrSetEc2Client(t *testing.T) {
	mgr := ExporterMgr{}
	mock2ec2 := &mockEc2Client{}
	mgr.SetEc2Client(mock2ec2)
	if diff := deep.Equal(mgr.Ec2Client(), mock2ec2); diff != nil {
		t.Errorf("Expected Ec2Client to give back what it got and it didn't: %#v", mgr.Ec2Client())
	}
}

func TestExporterMgrSetK8s(t *testing.T) {
	mgr := New(Config{})
	var mockk8s *mockK8s
	mgr.SetK8s(mockk8s)
	if diff := deep.Equal(mgr.K8s(), mockk8s); diff != nil {
		t.Error("Expected K8s to match what we set")
	}
}
func TestExporterMgrmapSrv(t *testing.T) {
	// Should succeed
	{
		srv := Service{
			SrvType: "Ec2",
			Srv: &Ec2{
				Tags: &[]Ec2Tag{
					{
						Tag:   "TestTag",
						Value: "TestValue",
					},
				},
			},
		}
		mgr := New(Config{})
		result, err := mgr.mapSrv("testservice", srv)
		if err != nil {
			t.Error(err)
		}
		if diff := deep.Equal(result, srv.Srv); diff != nil {
			t.Errorf("Expected to get an Ec2{} and didn't: %s", diff)
		}
	}
	// Bogus service should error
	{
		srv := Service{SrvType: "Bogus", Srv: &[]string{}}
		mgr := New(Config{})
		_, err := mgr.mapSrv("bogus", srv)
		if err == nil {
			t.Errorf("Expected mapSrv with bogus service to error and it didn't")
		}
	}
	// Bogus ec2 service
	{
		srv := Service{SrvType: "Ec2", Srv: &[]string{}}
		mgr := New(Config{})
		_, err := mgr.mapSrv("bogus", srv)
		if err == nil {
			t.Errorf("Expected mapSrv with Ec2 bogus service to error and it didn't")
		}
	}
}
func TestExporterMgrNew(t *testing.T) {
	mrg := New(Config{})
	mrgcompare := &ExporterMgr{Config: &Config{}}
	if diff := deep.Equal(mrg, mrgcompare); diff != nil {
		t.Errorf("Expected to get a new ExportMgr{} and didn't:\n%#v\n%#v,", mrg, mrgcompare)
	}
}
func TestExporterMgrRun(t *testing.T) {
	// Test not having to change anything
	{
		testcfg := Config{
			K8snamespace:           "default",
			K8sdeploytemplatespath: "../example-configs/k8stemplates/",
			K8slabels: map[string]string{
				"testlabel": "testvalue",
			},
			Services: map[string]interface{}{
				"prom-exporter-apache": &Service{
					SrvType: "Ec2",
					Srv: &Ec2{
						Tags: &[]Ec2Tag{
							{
								Tag:   "TestTag",
								Value: "TestValue",
							},
						},
					},
				},
			},
		}
		mgr := New(testcfg)
		mgr.SetEc2Client(&mockEc2Client{})
		mgr.SetK8s(&mockK8s{})
		err := mgr.Run()
		if err != nil {
			t.Error(err)
		}
	}
	// Test adding and deleting something
	{
		testcfg := Config{
			K8snamespace:           "default",
			K8sdeploytemplatespath: "../example-configs/k8stemplates/",
			K8slabels: map[string]string{
				"testlabel": "testvalue",
			},
			Services: map[string]interface{}{
				"prom-exporter-apache": &Service{
					SrvType: "Ec2",
					Srv: &Ec2{
						Tags: &[]Ec2Tag{
							{
								Tag:   "TestTag",
								Value: "TestValue",
							},
						},
					},
				},
			},
		}
		mgr := New(testcfg)
		mgr.SetEc2Client(&mockEc2Client{
			Instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								PrivateIpAddress: aws.String("172.16.0.2"),
								Tags: []*ec2.Tag{
									{
										Key:   aws.String("Name"),
										Value: aws.String("TestInstance2"),
									},
									{
										Key:   aws.String("tag:TestTag"),
										Value: aws.String("TestValue"),
									},
								},
							},
						},
					},
				},
			},
		})
		mgr.SetK8s(&mockK8s{})
		err := mgr.Run()
		if err != nil {
			t.Error(err)
		}
	}
}

func TestExporterMgrgregorianJoin(t *testing.T) {
	// Test for some aonly some bonly some both
	{
		a := []SrvInstance{
			{
				Name: "a",
				Addr: "a",
			},
			{
				Name: "b",
				Addr: "b",
			},
			{
				Name: "d",
				Addr: "d",
			},
			{
				Name: "e",
				Addr: "e",
			},
			{
				Name: "f",
				Addr: "f",
			},
			{
				Name: "g",
				Addr: "g",
			},
		}
		b := []SrvInstance{
			{
				Name: "b",
				Addr: "b",
			},
			{
				Name: "c",
				Addr: "c",
			},
		}
		aonly, bonly, both := gregorianJoin(a, b)
		if len(aonly) == 5 {
			for _, l := range []string{"a", "d", "e", "f", "g"} {
			Outter:
				for {
					for _, s := range aonly {
						if l == s.Name {
							break Outter
						}
					}
					t.Errorf("Didn't find %s in aonly", l)
				}
			}
		} else {
			t.Errorf("Expected 1 in aonly didn't get it: %#v", aonly)
		}
		if len(bonly) == 1 {
			if diff := deep.Equal(bonly[0], SrvInstance{Name: "c", Addr: "c"}); diff != nil {
				t.Errorf("Didn't get what we expected in bonly: %#v", diff)
			}
		} else {
			t.Errorf("Expected 1 in bonly didn't get it: %#v", bonly)
		}
		if len(both) == 1 {
			if diff := deep.Equal(both[0], SrvInstance{Name: "b", Addr: "b"}); diff != nil {
				t.Errorf("Didn't get what we expected in bonly: %#v", diff)
			}
		} else {
			t.Errorf("Expected 1 in both didn't get it: %#v", both)
		}
	}
	// Test for nothing in both
	{
		a := []SrvInstance{
			{
				Name: "a",
				Addr: "a",
			},
			{
				Name: "b",
				Addr: "b",
			},
			{
				Name: "c",
				Addr: "c",
			},
			{
				Name: "d",
				Addr: "d",
			},
		}
		b := []SrvInstance{
			{
				Name: "a",
				Addr: "a",
			},
			{
				Name: "b",
				Addr: "b",
			},
			{
				Name: "c",
				Addr: "c",
			},
			{
				Name: "d",
				Addr: "d",
			},
		}
		aonly, bonly, both := gregorianJoin(a, b)
		if len(aonly) > 0 {
			t.Errorf("Expected aonly to have 0 length: %#v", aonly)
		}
		if len(bonly) > 0 {
			t.Errorf("Expected bonly to have 0 length: %#v", bonly)
		}
		if diff := deep.Equal(both, a); diff != nil {
			t.Errorf("Didn't get what we expected in bonly: %#v", diff)
		}
	}
}

// Fake K8s Client
type mockK8s struct {
	ret *[]SrvInstance
}

func (m *mockK8s) Create(d *appsv1.Deployment) (*appsv1.Deployment, error) {
	return d, nil
}
func (m *mockK8s) Remove(string) error {
	return nil
}
func (m *mockK8s) Fetch() (*[]SrvInstance, error) {
	if m.ret != nil {
		return m.ret, nil
	}
	ret := &[]SrvInstance{
		{
			Name: "172.16.0.1",
			Addr: "172.16.0.1",
		},
	}
	return ret, nil
}
