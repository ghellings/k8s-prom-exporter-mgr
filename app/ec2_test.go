package exportermgr

import(
	"testing"
	"reflect"
	"fmt"
	"strings"

	"github.com/go-test/deep"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

)

func TestEc2SetEc2Client (t *testing.T) {
	ec2test := Ec2{}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ec2client := ec2.New(sess)
	ec2test.SetEc2Client(ec2client)
	checkec2client := ec2test.Ec2Client()
	if diff := deep.Equal(checkec2client, ec2client); diff != nil {
		t.Errorf("Expected ec2client to match set ec2client")
	}

}

func TestEc2CreateEc2Filters (t *testing.T) {
	ec2test := Ec2{
		Tags: &[]Ec2Tag {
			{
				Tag: "tag1",
				Value: "value1",
			},
			{
				Tag: "tag2",
				Value: "value2",
			},
		},
	}
	filter := []*ec2.Filter{
		{
			Name: aws.String("tag:tag1"),
			Values: []*string{aws.String("value1")},
		},
		{
			Name: aws.String("tag:tag2"),
			Values: []*string{aws.String("value2")},
		},
	}
	result := CreateEc2Filters(ec2test.Tags)
	if diff := deep.Equal(result,filter); diff != nil {
		t.Errorf("Expected createEc2Filters() to return []*ec2.Filter and it didn't: %#v", diff)
	}
}

func TestEc2Fetch (t *testing.T) { 
	ec2test := Ec2{
		Tags: &[]Ec2Tag{
			{
				Tag: "TestTag",
				Value: "TestValue",
			},
		},
		ec2client: &mockEc2Client{},
	}
	result,err := ec2test.Fetch()
	if err != nil {
		t.Errorf(err.Error())
	}
	if o := reflect.TypeOf(result); o != reflect.TypeOf(&[]SrvInstance{}) {
		t.Errorf("Expected to get a '*ec2.DescribeInstancesOutput' got: %s", o) 
	}
}

// Fake EC2 Client
type mockEc2Client struct {
	Instances *ec2.DescribeInstancesOutput
}

func (m *mockEc2Client) DescribeInstances(d *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	if m.Instances == nil {
		m.Instances = &ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances:  []*ec2.Instance{
					 	{
					  	PrivateIpAddress: aws.String("172.16.0.1"),
					  	Tags: []*ec2.Tag{
					  		{
					  	    Key: aws.String("Name"),
					  	    Value: aws.String("TestInstance"),
					  	  },
					  	  {
					  	    Key: aws.String("tag:TestTag"),
					  	    Value: aws.String("TestValue"),
					  	  },
					  	},
					  },
					},
				}, 
			},
		}
	}

	// Verify Tags match what we're planning on returning
	var tags []Ec2Tag
	for _,t := range m.Instances.Reservations[0].Instances[0].Tags {
		key := strings.TrimPrefix(*t.Key,"tag:")
		tag  := Ec2Tag{Tag: key, Value: *t.Value}
		tags = append(tags,tag)
	}
	filters := CreateEc2Filters(&tags)
	// Loop through array we created to ensure all items are found in what we're returning
	for _,df := range d.Filters {
		Found:
			for {
				for _,f := range filters {
					if d := deep.Equal(df,f); d == nil { break Found }
				}
				// Fake error that looks like Ec2 error
				return nil,fmt.Errorf("InvalidParameterValue: The filter '%s' is invalid\t\tstatus code: 400, request id: Fake-Error-ID",df)
			}
	}
	return m.Instances,nil
}