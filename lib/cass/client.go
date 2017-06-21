// Cassandra lib
package cass

import (
	"errors"
	"github.com/gocql/gocql"
)

// interface for exported functionality
type Client interface {
	// Users
	CreateUser(*User, *gocql.Batch) UpsertResult
	FetchUser(*User) (*User, error)
	// Beacons
	CreateBeacons([]*Beacon, *gocql.Batch) UpsertResult
	UpdateBeacons([]*Beacon, *gocql.Batch) UpsertResult
	FetchBeacon(*Beacon) (*Beacon, error)
	// Messages
	CreateMessage(*Message, *gocql.Batch) UpsertResult
	AddMessageDeployments(*Message, *[]string, *gocql.Batch) UpsertResult
	RemoveMessageDeployments(*Message, *[]string, *gocql.Batch) UpsertResult
	FetchMessage(*Message) (*Message, error)
	// Deployments
	PostDeploymentMetadata(*DeploymentMetadata, *gocql.Batch) UpsertResult
	PostDeployment(*Deployment) UpsertResult
}

type User struct {
	Id    *gocql.UUID `cql:"id"`
	Email string      `cql:"email"`
}

type Beacon struct {
	UserId     *gocql.UUID `cql:"user_id"`
	DeployName string      `cql:"deploy_name"`
	Name       string      `cql:"name"`
}

type Message struct {
	UserId      *gocql.UUID `cql:"user_id"`
	Name        string      `cql:"name"`
	Title       string      `cql:"title"`
	Url         string      `cql:"url"`
	Lang        string      `cql:"lang"`
	Deployments []string    `cql:"deployments"`
}

// Deployment is not an actual data structure stored in cassandra, but rather a construct that we disassemble into beacons. If a MessageName is provided, we will read/use that
// for settting a deployment method, otherwise creating a message if the Message field is set.
type Deployment struct {
	UserId      *gocql.UUID
	DeployName  string
	MessageName string
	Message     *Message
	BeaconNames []string
}

// DeploymentMetadata mirrors the need for a 'dashboard' level overview of beacon deployments.
type DeploymentMetadata struct {
	UserId      *gocql.UUID `cql:"user_id"`
	DeployName  string      `cql:"deploy_name"`
	MessageName string      `cql:"message_name"`
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

	return &CassClient{Sess: session}, nil
}

// Users ------------------------------------------------------------------------------

func (self *CassClient) CreateUser(u *User, batch *gocql.Batch) UpsertResult {
	uuid, _ := gocql.RandomUUID()
	template := `INSERT INTO users (id, email) VALUES (?, ?)`
	args := []interface{}{
		&uuid,
		u.Email,
	}

	if batch != nil {
		batch.Query(template, args...)
		return UpsertResult{Batch: batch, Err: nil}
	} else {
		return UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(template, args...).Exec(),
		}
	}

}

func (self *CassClient) FetchUser(u *User) (*User, error) {
	// instantiate user struct for unmarshalling
	matchedUser := &User{}
	var err error
	if u.Id != nil {
		err = self.Sess.Query(`SELECT id, email FROM users WHERE id = ?`, u.Id).Scan(&matchedUser.Id, &matchedUser.Email)
	} else {
		err = self.Sess.Query(`SELECT id, email FROM users_by_email WHERE email = ?`, u.Email).Scan(&matchedUser.Id, &matchedUser.Email)
	}

	if err != nil {
		return nil, err
	}

	if matchedUser.Id != nil {
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

		batch.Query(template, cmd...)
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
			bkn.DeployName,
			bkn.UserId,
			bkn.Name,
		}

		batch.Query(template, cmd...)
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

	template := `SELECT deploy_name FROM beacons WHERE user_id = ? AND name = ?`
	cmd := []interface{}{
		bkn.UserId,
		bkn.Name,
	}

	err := self.Sess.Query(template, cmd...).Scan(&resBkn.DeployName)
	return &resBkn, err
}

// Messages ------------------------------------------------------------------------------

func (self *CassClient) CreateMessage(m *Message, batch *gocql.Batch) UpsertResult {
	template := `INSERT INTO messages (user_id, name, title, url, lang, deployments) VALUES (?, ?, ?, ?, ?, ?)`
	args := []interface{}{
		m.UserId,
		m.Name,
		m.Title,
		m.Url,
		m.Lang,
		m.Deployments,
	}

	if batch != nil {
		batch.Query(template, args...)
		return UpsertResult{Batch: batch, Err: nil}
	} else {
		return UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(template, args...).Exec(),
		}
	}
}

func (self *CassClient) AddMessageDeployments(m *Message, additions *[]string, batch *gocql.Batch) UpsertResult {
	return self.addOrRemoveMessageDeployments(m, additions, true, batch)
}
func (self *CassClient) RemoveMessageDeployments(m *Message, removals *[]string, batch *gocql.Batch) UpsertResult {
	return self.addOrRemoveMessageDeployments(m, removals, false, batch)
}

// addOrRemoveMessageDeployments is the underlying function behind the exported AddMessageDeployments and RemoveMessageDeployments
func (self *CassClient) addOrRemoveMessageDeployments(m *Message, changes *[]string, add bool, batch *gocql.Batch) UpsertResult {
	if len(*changes) == 0 {
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
		changes,
		m.UserId,
		m.Name,
	}

	if batch != nil {
		batch.Query(template, args...)
		return UpsertResult{Batch: batch, Err: nil}
	} else {
		return UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(template, args...).Exec(),
		}
	}

}

func (self *CassClient) FetchMessage(m *Message) (*Message, error) {
	resMsg := &Message{}
	template := `SELECT user_id, name, title, url, lang, deployments FROM messages WHERE user_id = ? AND name = ?`
	args := []interface{}{
		m.UserId,
		m.Name,
	}

	err := self.Sess.Query(template, args...).Scan(&resMsg.UserId, &resMsg.Name, &resMsg.Title, &resMsg.Url, &resMsg.Lang, &resMsg.Deployments)
	return resMsg, err
}

// DeploymentMetadata
func (self *CassClient) PostDeploymentMetadata(md *DeploymentMetadata, batch *gocql.Batch) UpsertResult {
	template := `INSERT INTO deployments_metadata (user_id, deploy_name, message_name) VALUES (?, ?, ?)`
	args := []interface{}{
		md.UserId,
		md.DeployName,
		md.MessageName,
	}

	if batch != nil {
		batch.Query(template, args...)
		return UpsertResult{Batch: batch, Err: nil}
	} else {
		return UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(template, args...).Exec(),
		}
	}

}

// Deployments ------------------------------------------------------------------------------

// PostDeployment writes the current deployname to every beacon in the list via an update clause, causing any now-invalid records in the deployments materialized view (on top of beacons) to be deleted via a partition drop (& subsequently recreated)
// Must write Deployment to every beacon, deploy_name to any message used, and bNames/messageName/deployName to deployments_metadata
func (self *CassClient) PostDeployment(deployment *Deployment) UpsertResult {
	batch := gocql.NewBatch(gocql.LoggedBatch)
	// handle MessageName or Message fields appropriately
	if deployment.MessageName != "" {

		// Need to UPDATE the message with the new deployment_name (append to set)
		additions := []string{deployment.DeployName}
		self.AddMessageDeployments(&Message{UserId: deployment.UserId, Name: deployment.MessageName}, &additions, batch)
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
			Name:       bName,
		}
		bkns = append(bkns, &bkn)
	}

	self.UpdateBeacons(bkns, batch)

	var mName string
	if deployment.MessageName != "" {
		mName = deployment.MessageName
	} else {
		mName = deployment.Message.Name
	}
	// update metadata
	deploymentMeta := DeploymentMetadata{
		UserId:     deployment.UserId,
		DeployName: deployment.DeployName,
		// message could be provided as a reference (MessageName) or as a new Message object
		MessageName: mName,
	}

	self.PostDeploymentMetadata(&deploymentMeta, batch)

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
