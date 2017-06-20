// Cassandra lib
package cass

import (
	"errors"
	"github.com/gocql/gocql"
)

// interface for exported functionality
type Client interface {
	// Users
	CreateUser(u *User, batch *gocql.Batch) UpsertResult
	FetchUser(u *User) (*User, error)
	// Beacons
	CreateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult
	UpdateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult
	FetchBeacon(bkn *Beacon) (*Beacon, error)
	// Messages
	CreateMessage(m *Message, batch *gocql.Batch) UpsertResult
	AddMessageDeployments(m *Message, additions *Message.Deployments, batch *gocql.Batch) UpsertResult
	RemoveMessageDeployments(m *Message, removals *Message.Deployments, batch *gocql.Batch) UpsertResult
	addOrRemoveMessageDeployments(m *Message, changes *Message.Deployments, add bool, batch *gocql.Batch) UpsertResult
	FetchMessage(m *Message) (*Message, error)
	// Deployments
	PostDeploymentMetadata(md *DeploymentMetadata, batch *gocql.Batch) UpsertResult
	PostDeployment(deployment *Deployment) UpsertResult
}

type User struct {
	Id    gocql.UUID `cql:"id"`
	Email string     `cql:"email"`
}

type Beacon struct {
	UserId     gocql.UUID `cql:"user_id"`
	DeployName string     `cql:"deploy_name"`
	Name       string     `cql:"name"`
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
		err := self.Sess.Query(`SELECT id, email FROM users WHERE id = ?`, u.Id).Scan(&matchedUser.Id, &matchedUser.Email)
	} else {
		err := self.Sess.Query(`SELECT id, email FROM users_by_email WHERE email = ?`, u.Email).Scan(&matchedUser.Id, &matchedUser.Email)
	}

	if matchedUser.Id {
		return matchedUser, nil
	} else {
		return nil, errors.New("no matched user")
	}
}

// Beacons ------------------------------------------------------------------------------

func (self *CassClient) CreateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult {
	template := `INSERT INTO beacons (user_id, name, deploy_name) VALUES (?, ?, ?)`

	providedBatch := (batch != nil)
	if !providedBatch {
		batch = gocql.NewBatch(gocql.LoggedBatch)
	}

	res := UpsertResult{
		Batch: batch,
	}

	for _, bkn := range beacons {
		cmd := []interface{}{
			template,
			bkn.UserId,
			bkn.Name,
			bkn.DeployName,
		}

		batch.Query(cmd...)
	}

	// If a batch was provided, we do not need to execute the query, it may be done as part of a later transaction.
	if !providedBatch {
		res.Err = self.Sess.ExecuteBatch(batch)
	}

	return res
}

// UpdateBeacons must use an if exists clause to prevent errors like inserting a beacon which a user does not own.
func (self *CassClient) UpdateBeacons(beacons []*Beacon, batch *gocql.Batch) UpsertResult {
	template := `UPDATE beacons SET deploy_name = ? WHERE user_id = ? AND name = ? IF EXISTS`

	providedBatch := (batch != nil)
	if !providedBatch {
		batch = gocql.NewBatch(gocql.LoggedBatch)
	}

	res := UpsertResult{
		Batch: batch,
	}

	for _, bkn := range beacons {
		cmd := []interface{}{
			template,
			bkn.DeployName,
			bkn.UserId,
			bkn.Name,
		}

		batch.Query(cmd...)
	}

	// If a batch was provided, we do not need to execute the query, it may be done as part of a later transaction.
	if !providedBatch {
		res.Err = self.Sess.ExecuteBatch(batch)
	}

	return res

}

// FetchBeacons takes a slice of Beacons with primary keys defined, fetches the remaining data, & updates the structs
func (self *CassClient) FetchBeacon(bkn *Beacon) (*Beacon, error) {
	resBkn := Beacon{
		UserId: bkn.UserId,
		Name:   bkn.Name,
	}

	cmd := []interface{}{
		`SELECT deploy_name FROM beacons WHERE user_id = ? AND name = ?`,
		bkn.UserId,
		bkn.Name,
	}

	err := self.Sess.Query(cmd...).Scan(&resBkn.DeployName)
	return &resBkn, err
}

// Messages ------------------------------------------------------------------------------

func (self *CassClient) CreateMessage(m *Message, batch *gocql.Batch) UpsertResult {
	template := `INSERT INTO messages (user_id, name, title, url, lang, deployments) VALUES (?, ?, ?, ?, ?, ?)`
	args := []interface{}{
		template,
		m.UserId,
		m.Name,
		m.Title,
		m.Url,
		m.Lang,
		m.Deployments,
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

func (self *CassClient) AddMessageDeployments(m *Message, additions *Message.Deployments, batch *gocql.Batch) UpsertResult {
	return self.addOrRemoveMessageDeployments(m, additions, true, batch)
}
func (self *CassClient) RemoveMessageDeployments(m *Message, removals *Message.Deployments, batch *gocql.Batch) UpsertResult {
	return self.AddMessageDeployments(m, removals, false, batch)
}

// addOrRemoveMessageDeployments is the underlying function behind the exported AddMessageDeployments and RemoveMessageDeployments
func (self *CassClient) addOrRemoveMessageDeployments(m *Message, changes *Message.Deployments, add bool, batch *gocql.Batch) UpsertResult {
	if len(changes) == 0 {
		return UpsertResult{Err: errors.New("must specify changes to message deployments")}
	}

	var operator string
	if add {
		operator = "+"
	} else {
		operator = "-"
	}

	template := `UPDATE messages SET deployments = deployments ` + operator + `? WHERE user_id = ? AND name = ? IF EXISTS`
	args := []interface{}{
		template,
		changes,
		m.UserId,
		m.Name,
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

func (self *CassClient) FetchMessage(m *Message) (*Message, error) {
	resMsg := Message{}
	args := []interface{}{
		`SELECT user_id, name, title, url, lang, deployments FROM messages WHERE user_id = ? AND name = ?`,
		m.UserId,
		m.Name,
	}

	err := self.Sess.Query(args...).Scan(resMsg.UserId, resMsg.Name, resMsg.Title, resMsg.Url, resMsg.Lang, resMsg.Deployments)
	return resMsg, err
}

// DeploymentMetadata
func (self *CassClient) PostDeploymentMetadata(md *DeploymentMetadata, batch *gocql.Batch) UpsertResult {
	template := `INSERT INTO deployments_metadata (user_id, deploy_name, message_name) VALUES (?, ?, ?)`
	args := []interface{}{
		template,
		md.UserId,
		md.DeployName,
		md.MessageName,
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
