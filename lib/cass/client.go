// Cassandra lib
package cass

import (
	"errors"
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
	Email string     `cql:"email"`
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

// DeploymentMetadata mirrors the need for a 'dashboard' level overview of beacon deployments.
type DeploymentMetadata struct {
	UserId      gocql.UUID `cql:"user_id"`
	DeployName  string     `cql:"deploy_name"`
	MessageName string     `cql:"message_name"`
}

type CassClient struct {
	Sess *gocql.Session
}

// UpsertResult is a wrapper type, including a possible batch & error. It can be used as the return value for batched or unbatched DML statements
type UpsertResult struct {
	Batch *gocql.Batch
	Err   error
}

// Instantiation

func Connect(cluster *gocql.ClusterConfig, keyspace string) (*CassClient, error) {
	cluster.Keyspace = keyspace
	session, connectErr := cluster.CreateSession()
	if connectErr != nil {
		return nil, connectErr
	}

	return *CassClient{Sess: session}
}

// Users ------------------------------------------------------------------------------

func (self *CassClient) CreateUser(u *User, batch *gocql.Batch) UpsertResult {
	args := []interface{}{
		`INSERT INTO users (id, email) VALUES (?, ?)`,
		gocql.RandomUUID(),
		u.Email,
	}

	if batch != nil {
		batch.Query(args...)
		return UpsertResult{Batch: batch, Err: nil}
	} else {
		return UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(args...).Exec(),
		}
	}

}

func (self *CassClient) FetchUser(u *User) (*User, error) {
	// instantiate user struct for unmarshalling
	matchedUser := &User{}
	if u.Id {
		err := self.Sess.Query(`Select id, email FROM users WHERE id = ?`, u.Id).Scan(&matchedUser.Id, &matchedUser.Email)
	} else {
		err := self.Sess.Query(`Select id, email FROM users_by_email WHERE email = ?`, u.Email).Scan(&matchedUser.Id, &matchedUser.Email)
	}

	if matchedUser.Id {
		return matchedUser, nil
	} else {
		return nil, errors.New("no matched user")
	}
}

// Beacons ------------------------------------------------------------------------------

func (self *CassClient) CreateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult {}

// UpdateBeacons
// Note: UpdateBeacons must use an if exists clause to prevent errors like inserting a beacon which a user does not own.
func (self *CassClient) UpdateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult {}
func (self *CassClient) FetchBeacons(beacons []*Beacon) (*Beacon, error)                  {}

// Messages ------------------------------------------------------------------------------

func (self *CassClient) CreateMessage(m *Message, batch *gocql.Batch) UpsertResult {}
func (self *CassClient) ChangeMessageDeployments(m *Message, additions *Message.Deployments, removals *Message.Deployments, batch *gocql.Batch) UpsertResult {
}
func (self *CassClient) FetchMessage(m *Message, batch *gocql.Batch) (*Message, error) {}

// DeploymentMetadata
func (self *CassClient) PostDeploymentMetadata(md *DeploymentMetadata, batch *gocql.Batch) UpsertResult {
}

// Deployments ------------------------------------------------------------------------------

// PostDeployment writes the current deployname to every beacon in the list via an update clause, causing any now-invalid records in the deployments materialized view (on top of beacons) to be deleted via a partition drop (& subsequently recreated)
// Must write Deployment to every beacon, deploy_name to any message used, and bNames/messageName/deployName to deployments_metadata
func (self *CassClient) PostDeployment(deployment *Deployment) UpsertResult {
	batch := gocql.NewBatch(gocql.LoggedBatch)
	// handle MessageName or Message fields appropriately
	if deployment.MessageName {
		// fetch message from id
		msg, fetchErr := self.FetchMessage(&Message{Name: deployment.MessageName, UserId: deployment.UserId})
		if fetchErr != nil {
			return nil, fetchErr
		}

		// Need to UPDATE the message with the new deployment_name (append to set)
		additions := Message.Deployments{deployment.DeployName}
		self.ChangeMessageDeployments(&Message{UserId: deployment.UserId, Name: deployment.MessageName}, &additions, nil, batch)
	} else {
		// Otherwise, create a new message from the provided Message
		deployment.Message.Deployments = []string{deployment.DeployName}
		deployment.Message.UserId = deployment.UserId
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

	// update metadata
	deploymentMeta := DeploymentMetadata{
		UserId:     deployment.UserId,
		DeployName: deployment.DeployName,
		// message could be provided as a reference (MessageName) or as a new Message object
		MessageName: deployment.MessageName || deployment.Message.Name,
	}

	self.PostDeploymentMetadata(deploymentMeta, batch)

	//Execute batch w/ context & return res
	res := UpsertResult{
		Err:   self.Sess.ExecuteBatch(batch),
		Batch: batch,
	}

	return res
}

// FetchDeploymentBeacons uses the deployments materialized view to gather a list of beacons associated with a deployment.
// func FetchDeploymentBeacons() {}

// FetchDeploymentMetadata fetches the metadata record for a deployment
// func FetchDeploymentMetadata() {}
