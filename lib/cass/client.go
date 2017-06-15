// Cassandra lib
package cass

import ()

// interface for exported functionality
type Client interface {
	// Deployments
	// Beacons
	// Users
	// Messages
	// Schedule
}

type Message struct {
	UserId int
	Name   string
	Title  string
	Url    string
	Lang   string
}

type Deployment struct {
	UserId      uuid
	DeployName  string
	Message     *Message
	BeaconNames []string
}

type Beacon struct {
	UserId      uuid
	DeployName  string
	MessageName string
}

type CassClient struct{}

// UpsertDeployment writes the current deployname to every beacon in the list via an update clause, causing any now-invalid records in the deployments materialized view (on top of beacons) to be deleted via a partition drop (& subsequently recreated)
// Must write Deployment to every beacon
func (self *CassClient) UpsertDeployment() {
	// take list of beacon names, write the new deployname to them all

	// if message
}

// FetchDeployment uses the deployments materialized view to gather a list of beacons associated with a deployment.
// func FetchDeployment() {}

// Crud ops for messages

type User struct {
	id    uuid
	email string
}
