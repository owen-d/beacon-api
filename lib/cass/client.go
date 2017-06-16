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
	UserId     gocql.UUID `cql:"user_id"`
	DeployName string     `cql:"deploy_name"`
	BeaconName string     `cql:"beacon_name"`
}

type Message struct {
	UserId      gocql.UUID `cql:"user_id"`
	Name        string     `cql:"name"`
	Title       string     `cql:"title"`
	Url         string     `cql:"url"`
	Lang        string     `cql:"lang"`
	Deployments []string   `cql:"deployments"`
}

// Deployment is not an actual data structure stored in cassandra, but rather a construct that we disassemble into beacons. If a MessageName is provided, we will read/use that
// for settting a deployment method, otherwise creating a message if the Message field is set.
type Deployment struct {
	UserId      gocql.UUID
	DeployName  string
	MessageName string
	Message     *Message
	BeaconNames []string
}

type CassClient struct{}

// UpsertResult is a wrapper type, including a possible batch & error. It can be used as the return value for batched or unbatched DML statements
type UpsertResult struct {
	Batch *gocql.Batch
	err   error
}

// Users ------------------------------------------------------------------------------

func (self *CassClient) CreateUser(u *User, batch *gocql.Batch) UpsertResult {}
func (self *CassClient) FetchUser(u *User, batch *gocql.Batch)               {}

// Beacons ------------------------------------------------------------------------------

func (self *CassClient) CreateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult {}
func (self *CassClient) UpdateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult {}
func (self *CassClient) FetchBeacons(beacons []*Beacon) (*Beacon, error)                  {}

// Messages ------------------------------------------------------------------------------

func (self *CassClient) CreateMessage(m *Message, batch *gocql.Batch) UpsertResult {}
func (self *CassClient) ChangeMessageDeployments(m *Message, additions *Message.Deployments, removals *Message.Deployments, batch *gocql.Batch) UpsertResult {
}
func (self *CassClient) FetchMessage(m *Message, batch *gocql.Batch) (*Message, error) {}

// Deployments ------------------------------------------------------------------------------

// PostDeployment writes the current deployname to every beacon in the list via an update clause, causing any now-invalid records in the deployments materialized view (on top of beacons) to be deleted via a partition drop (& subsequently recreated)
// Must write Deployment to every beacon
func (self *CassClient) PostDeployment(deployment *Deployment) UpsertResult {
	batch := gocql.NewBatch(gocql.LoggedBatch)
	// Note, should handle all these in batch.
	// handle MessageName or Message fields appropriately
	if deployment.MessageName {
		// fetch message from id
		msg, fetchErr := self.FetchMessage(&Message{Name: deployment.MessageName})
		if fetchErr != nil {
			return nil, fetchErr
		}

		// assign message into deployment
		deployment.Message = msg
	} else {
		// Otherwise, create a new message from the provided Message
		deployment.Message.Deployments = []string{deployment.DeployName}
		self.CreateMessage(deployment.Message, batch)

	}
	// take list of beacon names, write the new deployname to them all
	bkns := make([]*Beacon, len(deployment.BeaconNames))
	for _, bName := range deployment.BeaconNames {
		bkn := Beacon{
			UserId:     deployment.UserId,
			DeployName: deployment.DeployName,
			BeaconName: bName,
		}
		bkns = append(bkns, bkn)
	}

	self.UpdateBeacons(bkns, batch)

	//Execute batch w/ context & return res
}

// FetchDeployment uses the deployments materialized view to gather a list of beacons associated with a deployment.
// func FetchDeployment() {}
