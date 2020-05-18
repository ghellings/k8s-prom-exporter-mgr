package exportermgr

import(
	"fmt"
	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

type Ec2 struct {
	Tags *[]Ec2Tag
	ec2client ec2clientinterface
	name string
}

type Ec2Tag struct {
	Tag string
	Value string 
}

type ec2clientinterface interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

func (e *Ec2) SetEc2Client(ec2client ec2clientinterface) {
	e.ec2client = ec2client
}

func (e *Ec2) Ec2Client() ec2clientinterface {
	if e.ec2client == nil {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
	  	SharedConfigState: session.SharedConfigEnable,
		}))
		ec2client := ec2.New(sess)
		e.SetEc2Client(ec2client)
	}
	return e.ec2client
}

func(e *Ec2) SetName(name string) {
	e.name = name
}

func(e *Ec2) Name() string {
	return e.name
}

// Fetch
func (e *Ec2) Fetch() (*[]SrvInstance, error) {
	ec2client := e.Ec2Client()
	params := &ec2.DescribeInstancesInput{
		Filters: CreateEc2Filters(e.Tags),
	}
	log.Tracef("Looking for Ec2 instances with params: %#v", params)
	instancesoutput, err := ec2client.DescribeInstances(params)
	if err != nil {
		return nil, err
	}
	var srvinstances []SrvInstance
	for _, r := range instancesoutput.Reservations {
		for _, i := range r.Instances {
			instance := SrvInstance{
				Name: fmt.Sprintf("%s-%s",e.Name(),*i.PrivateIpAddress),
				Addr: *i.PrivateIpAddress,
			}
			srvinstances = append(srvinstances, instance)
		}
	}
	return &srvinstances,err
}

//  Create filters Ec2 client wants for describing instances
func CreateEc2Filters(e *[]Ec2Tag) []*ec2.Filter {
	var ec2filters []*ec2.Filter
	for _, f := range *e {
		filter := &ec2.Filter{
			Name: aws.String("tag:"+f.Tag),
			Values: []*string{aws.String(f.Value)},
		}
		ec2filters = append(ec2filters, filter)
	}
	return ec2filters
}