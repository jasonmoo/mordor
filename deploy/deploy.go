package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/jasonmoo/goamz/aws"
	"github.com/jasonmoo/goamz/ec2"

	"github.com/jasonmoo/pd"
)

const (
	DefaultConcurrency = 4

	Instance_Name             = "mordor-test"
	Instance_ClassName        = "mordor-test"
	Instance_Ami              = "ami-f34032c3" // ubuntu 14.04 default
	Instance_Type             = "m3.large"
	Instance_AvailabilityZone = "us-west-2b"

	Instance_User    = "ubuntu"
	Instance_KeyName = "mordor"
	Instance_PubKey  = "mordor.pem"
)

var (
	Instance_Region         = aws.USWest2
	Instance_SecurityGroups = []ec2.SecurityGroup{
		ec2.SecurityGroup{Name: "mordor"},
	}

	// global ec2 region object
	region *ec2.EC2

	// global pool of actionable servers
	pool *pd.Pool

	// default actions
	server_names = flag.Bool("server_dns", false, "output the servers")
	run          = flag.String("run", "", "command to run on all nodes")
	setup        = flag.Bool("setup", false, "setup server with required packages")
	deploy       = flag.Bool("deploy", false, "deploy the binary/confs/deps")
	add_node     = flag.Int("add_node", 0, "add a node to current config")
	remove_node  = flag.Int("remove_node", 0, "remove a node from current pool")

	instance_type     = flag.String("instance_type", Instance_Type, "instance type to launch")
	availability_zone = flag.String("availability_zone", Instance_AvailabilityZone, "az to operate on")
)

func init() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	region = ec2.New(aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}, Instance_Region)

	log.Println("pd starting up...")
	pd.PrintInstances(region, Instance_ClassName)

	servers := pd.GetPublicDNS(region, Instance_ClassName)

	pool = pd.NewPool(Instance_User, Instance_PubKey, servers, DefaultConcurrency)

}

func main() {

	defer log.Println("done.")

	switch {

	case *server_names:
		for _, server := range pool.Servers {
			fmt.Println(server)
		}

	case *run != "":
		pool.WaitForPort(22)
		pd.Must(pool.Run(*run))

	case *setup:
		log.Println("Setting up...")
		// pd.OpenPortRange(region, Instance_ClassName, 1025, 65535)

		pool.WaitForPort(22)
		pd.MustShort(pool.Sudo("iptables -t nat -A PREROUTING -p tcp --dport 1025:65535 -j REDIRECT --to-ports 10000"))

	case *deploy:
		// build the target binary
		pd.BuildGoBinary("$GOPATH/src/github.com/jasonmoo/mordor/listen", "listen", "linux", "amd64")

		// push the new conf
		pool.WaitForPort(22)
		pd.Must(pool.Sudo("chown -R %s /etc/init", Instance_User))
		pd.Must(pool.Rsync("-az", "listen.conf", "/etc/init/listen.conf"))

		// push the new binary
		pd.Must(pool.Rsync("-az", "$GOPATH/src/github.com/jasonmoo/mordor/listen/listen", "listen"))

		// swap the new binary with the old and restart
		// touch ensures file on first deploy
		pd.Must(pool.Sudo("restart listen || start listen"))

	case *add_node > 0:
		pd.AddNode(region, Instance_Name, Instance_ClassName, &ec2.RunInstancesOptions{
			ImageId:          Instance_Ami,
			MinCount:         *add_node,
			MaxCount:         *add_node,
			AvailabilityZone: *availability_zone,
			KeyName:          Instance_KeyName,
			InstanceType:     *instance_type,
			SecurityGroups:   Instance_SecurityGroups,
		})
		log.Println("Sleeping for 30s to ensure boot...")
		time.Sleep(30 * time.Second)

	case *remove_node > 0:
		pd.RemoveNode(region, Instance_ClassName, *remove_node)

	default:
		fmt.Println("Usage: ")
		flag.PrintDefaults()

	}

}
