// Cassandra lib
package cass

import (
	"github.com/gocql/gocql"
)

// interface for exported functionality
type Client interface {
	// Deployments
	// Beacons
	// Users
	// Messages
	// Schedule
}

type User struct {
	Id    gocql.UUID `cql:"id"`
	email string     `cql:"email"`
}

type Beacon struct {
	UserId      gocql.UUID `cql:"user_id"`
	DeployName  string     `cql:"deploy_name"`
	MessageName string     `cql:"message_name"`
}

type Message struct {
	UserId      gocql.UUID `cql:"user_id"`
	Name        string     `cql:"name"`
	Title       string     `cql:"title"`
	Url         string     `cql:"url"`
	Lang        string     `cql:"lang"`
	Deployments []string   `cql:"deployments"`
}

// Deployment is not an actual data structure stored in cassandra, but rather a construct that we disassemble into beacons. If a MessageId is provided, we will read/use that
// for settting a deployment method, otherwise creating a message if the Message field is set.
type Deployment struct {
	UserId      gocql.UUID
	DeployName  string
	MessageId   string
	Message     *Message
	BeaconNames []string
}

type CassClient struct{}

// Users ------------------------------------------------------------------------------

func (self *CassClient) CreateUser(u *User) {}
func (self *CassClient) FetchUser(u *User)  {}

// Beacons ------------------------------------------------------------------------------

func (self *CassClient) CreateBeacons(beacons []*Beacon) {}
func (self *CassClient) UpdateBeacons(beacons []*Beacon) {}

// Messages ------------------------------------------------------------------------------

func (self *CassClient) CreateMessage(m *Message) {}
func (self *CassClient) ChangeMessageDeployments(m *Message, additions *Message.Deployments, removals *Message.Deployments) {
}

// Deployments ------------------------------------------------------------------------------

// PostDeployment writes the current deployname to every beacon in the list via an update clause, causing any now-invalid records in the deployments materialized view (on top of beacons) to be deleted via a partition drop (& subsequently recreated)
// Must write Deployment to every beacon
func (self *CassClient) PostDeployment(deployment *Deployment) {
	// handle MessageId or Message fields appropriately
	if deployment.MessageId {
		// fetch message from id
	} else {
		// create message from msg struct
	}
	// take list of beacon names, write the new deployname to them all
	for _, bName := range deployment.BeaconNames {

	}

}

// FetchDeployment uses the deployments materialized view to gather a list of beacons associated with a deployment.
// func FetchDeployment() {}
