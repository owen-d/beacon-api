// Cassandra lib
package cass

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/gocql/gocql"
)

// interface for exported functionality
type Client interface {
	// Users
	CreateUser(*User, *gocql.Batch) *UpsertResult
	FetchUser(*User) (*User, error)
	// Beacons
	CreateBeacons([]*Beacon, *gocql.Batch) *UpsertResult
	UpdateBeacons([]*Beacon) *UpsertResult
	FetchBeacon(*Beacon) (*Beacon, error)
	FetchUserBeacons(*gocql.UUID) ([]*Beacon, error)
	// Messages
	CreateMessage(*Message, *gocql.Batch) *UpsertResult
	AddMessageDeployments(*Message, []string, *gocql.Batch) *UpsertResult
	RemoveMessageDeployments(*Message, []string, *gocql.Batch) *UpsertResult
	FetchMessage(*Message) (*Message, error)
	FetchMessages(*gocql.UUID, uint8) ([]*Message, error)
	// Deployments
	FetchDeployment(*Deployment) (*Deployment, error)
	PostDeployment(*Deployment) *UpsertResult
	FetchDeploymentBeacons(dep *Deployment) ([]*Beacon, error)
	// Metadata
	FetchDeploymentsMetadata(*gocql.UUID) ([]*Deployment, error)
	FetchDeploymentMetadata(*gocql.UUID, string) (*Deployment, error)
	PostDeploymentMetadata(*Deployment, *gocql.Batch) *UpsertResult
}

const (
	DefaultLimit = 250
)

type User struct {
	Id    *gocql.UUID `cql:"id"`
	Email string      `cql:"email"`
}

type Beacon struct {
	UserId     *gocql.UUID `cql:"user_id" json:"user_id"`
	DeployName string      `cql:"deploy_name" json:"deploy_name"`
	Name       []byte      `cql:"name"`
}

func (self *Beacon) MarshalJSON() ([]byte, error) {
	type Alias Beacon
	return json.Marshal(&struct {
		Name string `json:"name"`
		*Alias
	}{
		Name:  hex.EncodeToString(self.Name),
		Alias: (*Alias)(self),
	})
}

func (self *Beacon) UnmarshalJSON(data []byte) error {
	type Alias Beacon
	aux := struct {
		Name string `json:name`
		*Alias
	}{
		Alias: (*Alias)(self),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var decodeErr error
	if self.Name, decodeErr = hex.DecodeString(aux.Name); decodeErr != nil {
		return decodeErr
	}
	return nil
}

type Message struct {
	UserId      *gocql.UUID `cql:"user_id" json:-`
	Name        string      `cql:"name" json:"name"`
	Title       string      `cql:"title" json:"title"`
	Url         string      `cql:"url" json:"url"`
	Lang        string      `cql:"lang" json:"lang"`
	Deployments []string    `cql:"deployments" json:"deployments"`
}

// Deployment is not an actual data structure stored in cassandra, but rather a construct that we disassemble into beacons. If a MessageName is provided, we will read/use that
// for settting a deployment method, otherwise creating a message if the Message field is set.
type Deployment struct {
	UserId      *gocql.UUID `json:"user_id"`
	DeployName  string      `json:"name"`
	MessageName string      `json:"message_name,omitempty"`
	Message     *Message    `json:"message,omitempty"`
	BeaconNames [][]byte    `json:"beacon_names"`
}

func (self *Deployment) MarshalJSON() ([]byte, error) {
	type Alias Deployment

	bNames := make([]string, len(self.BeaconNames))

	for i, name := range self.BeaconNames {
		bNames[i] = hex.EncodeToString(name)
	}
	return json.Marshal(&struct {
		BeaconNames []string `json:"beacon_names"`
		*Alias
	}{
		BeaconNames: bNames,
		Alias:       (*Alias)(self),
	})
}

func (self *Deployment) UnmarshalJSON(data []byte) error {
	type Alias Deployment
	aux := struct {
		BeaconNames []string `json:beacon_names`
		*Alias
	}{
		Alias: (*Alias)(self),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	bNames := make([][]byte, len(aux.BeaconNames))
	for i, name := range aux.BeaconNames {
		if decoded, decodeErr := hex.DecodeString(name); decodeErr != nil {
			return decodeErr
		} else {
			bNames[i] = decoded
		}
	}
	self.BeaconNames = bNames
	return nil
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

func (self *CassClient) CreateUser(u *User, batch *gocql.Batch) *UpsertResult {
	uuid, _ := gocql.RandomUUID()
	template := `INSERT INTO users (id, email) VALUES (?, ?)`
	args := []interface{}{
		&uuid,
		u.Email,
	}

	if batch != nil {
		batch.Query(template, args...)
		return &UpsertResult{Batch: batch, Err: nil}
	} else {
		return &UpsertResult{
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

func (self *CassClient) CreateBeacons(beacons []*Beacon, batch *gocql.Batch) *UpsertResult {
	template := `INSERT INTO beacons (user_id, name, deploy_name) VALUES (?, ?, ?) IF NOT EXISTS`

	providedBatch := (batch != nil)
	if !providedBatch {
		batch = gocql.NewBatch(gocql.LoggedBatch)
	}

	res := UpsertResult{
		Batch: batch,
	}

	for _, bkn := range beacons {
		cmd := []interface{}{
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

	return &res
}

// UpdateBeacons must use an if exists clause to prevent errors like inserting a beacon which a user does not own.
func (self *CassClient) UpdateBeacons(beacons []*Beacon) *UpsertResult {
	template := `UPDATE beacons SET deploy_name = ? WHERE user_id = ? AND name = ? IF EXISTS`
	dispatch := newDispatcher()

	for _, bkn := range beacons {
		cmd := []interface{}{
			bkn.DeployName,
			bkn.UserId,
			bkn.Name,
		}

		dispatch.Register(func() *UpsertResult {
			return &UpsertResult{
				Batch: nil,
				Err:   self.Sess.Query(template, cmd...).Exec(),
			}
		})
	}

	for i := uint32(0); i < dispatch.Ct; i++ {
		res := <-dispatch.Ch

		if res.Err != nil {
			return res
		}
	}

	res := UpsertResult{
		Err: nil,
		// return nil batch b/c theres no collective batch
		Batch: nil,
	}

	return &res

}

// FetchBeacon takes a slice of Beacons with primary keys defined, fetches the remaining data, & updates the structs
func (self *CassClient) FetchBeacon(bkn *Beacon) (*Beacon, error) {
	resBkn := Beacon{
		UserId: bkn.UserId,
		Name:   bkn.Name,
	}

	template := `SELECT user_id, deploy_name FROM beacons WHERE user_id = ? AND name = ?`
	cmd := []interface{}{
		bkn.UserId,
		bkn.Name,
	}

	err := self.Sess.Query(template, cmd...).Scan(&resBkn.UserId, &resBkn.DeployName)
	return &resBkn, err
}

// FetchUserBeacons returns a slice of beacons belonging to a user
func (self *CassClient) FetchUserBeacons(userId *gocql.UUID) ([]*Beacon, error) {
	template := `SELECT user_id, deploy_name, name FROM beacons WHERE user_id = ? LIMIT ?`
	args := []interface{}{
		userId,
		DefaultLimit,
	}

	resRows := make([]*Beacon, 0)
	iter := self.Sess.Query(template, args...).Iter()
	shell := map[string]interface{}{}
	for iter.MapScan(shell) {
		id := shell["user_id"].(gocql.UUID)
		resRows = append(resRows, &Beacon{
			UserId:     &id,
			DeployName: shell["deploy_name"].(string),
			Name:       shell["name"].([]uint8),
		})

		// since shell is used in each iteration, we must clear it.
		shell = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return resRows, nil

}

// Messages ------------------------------------------------------------------------------

func (self *CassClient) CreateMessage(m *Message, batch *gocql.Batch) *UpsertResult {
	template := `INSERT INTO messages (user_id, name, title, url, lang, deployments) VALUES (?, ?, ?, ?, ?, ?) IF NOT EXISTS`
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
		return &UpsertResult{Batch: batch, Err: nil}
	} else {
		return &UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(template, args...).Exec(),
		}
	}
}

func (self *CassClient) AddMessageDeployments(m *Message, additions []string, batch *gocql.Batch) *UpsertResult {
	return self.addOrRemoveMessageDeployments(m, additions, true, batch)
}
func (self *CassClient) RemoveMessageDeployments(m *Message, removals []string, batch *gocql.Batch) *UpsertResult {
	return self.addOrRemoveMessageDeployments(m, removals, false, batch)
}

// addOrRemoveMessageDeployments is the underlying function behind the exported AddMessageDeployments and RemoveMessageDeployments
func (self *CassClient) addOrRemoveMessageDeployments(m *Message, changes []string, add bool, batch *gocql.Batch) *UpsertResult {
	if len(changes) == 0 {
		return &UpsertResult{Err: errors.New("must specify changes to message deployments")}
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
		return &UpsertResult{Batch: batch, Err: nil}
	} else {
		return &UpsertResult{
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

func (self *CassClient) FetchMessages(id *gocql.UUID, lim uint8) ([]*Message, error) {
	template := `SELECT user_id, name, title, url, lang, deployments FROM messages WHERE user_id = ? LIMIT ?`
	args := []interface{}{
		id,
		lim,
	}

	resRows := make([]*Message, 0)
	iter := self.Sess.Query(template, args...).Iter()
	shell := map[string]interface{}{}
	for iter.MapScan(shell) {
		id := shell["user_id"].(gocql.UUID)
		resRows = append(resRows, &Message{
			UserId:      &id,
			Name:        shell["name"].(string),
			Title:       shell["title"].(string),
			Url:         shell["url"].(string),
			Lang:        shell["lang"].(string),
			Deployments: shell["deployments"].([]string),
		})

		// since shell is used in each iteration, we must clear it.
		shell = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return resRows, nil

}

// DeploymentMetadata
func (self *CassClient) PostDeploymentMetadata(dep *Deployment, batch *gocql.Batch) *UpsertResult {
	template := `INSERT INTO deployments_metadata (user_id, deploy_name, message_name) VALUES (?, ?, ?)`
	args := []interface{}{
		dep.UserId,
		dep.DeployName,
		dep.MessageName,
	}

	if batch != nil {
		batch.Query(template, args...)
		return &UpsertResult{Batch: batch, Err: nil}
	} else {
		return &UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(template, args...).Exec(),
		}
	}

}

// FetchDeploymentsMetadata
func (self *CassClient) FetchDeploymentsMetadata(userId *gocql.UUID) ([]*Deployment, error) {
	resRows := make([]*Deployment, 0)
	template := `SELECT user_id, deploy_name, message_name FROM deployments_metadata WHERE user_id = ? LIMIT ?`
	args := []interface{}{
		userId,
		DefaultLimit,
	}
	iter := self.Sess.Query(template, args...).Iter()
	shell := map[string]interface{}{}
	for iter.MapScan(shell) {
		id := shell["user_id"].(gocql.UUID)
		resRows = append(resRows, &Deployment{
			UserId:      &id,
			DeployName:  shell["deploy_name"].(string),
			MessageName: shell["message_name"].(string),
		})

		// since shell is used in each iteration, we must clear it.
		shell = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return resRows, nil
}

// FetchDeploymentMetadata is the single version of FetchDeploymentsMetadata. It requires a DeployName.
func (self *CassClient) FetchDeploymentMetadata(userId *gocql.UUID, depName string) (*Deployment, error) {
	res := &Deployment{}
	template := `SELECT user_id, deploy_name, message_name FROM deployments_metadata WHERE user_id = ? AND deploy_name = ? LIMIT 1`
	args := []interface{}{
		userId,
		depName,
	}
	err := self.Sess.Query(template, args...).Scan(&res.UserId, &res.DeployName, &res.MessageName)

	if err != nil {
		return nil, err
	}
	return res, nil
}

// Deployments ------------------------------------------------------------------------------

// PostDeployment writes the current deployname to every beacon in the list via an update clause, causing any now-invalid records in the deployments materialized view (on top of beacons) to be deleted via a partition drop (& subsequently recreated)
// Must write Deployment to every beacon, deploy_name to any message used, and bNames/messageName/deployName to deployments_metadata
func (self *CassClient) PostDeployment(deployment *Deployment) *UpsertResult {
	// execute everything concurrently & pass results through dispatcher
	dispatch := newDispatcher()

	// handle MessageName or Message fields appropriately
	if mName := deployment.MessageName; mName != "" {
		// must make sure MessageName matches an existing message & assign it into the deployment struct
		foundMsg, err := self.FetchMessage(&Message{UserId: deployment.UserId, Name: mName})
		if err != nil {
			return &UpsertResult{Err: err}
		} else {
			deployment.Message = foundMsg
		}

		// Need to UPDATE the message with the new deployment_name (append to set)
		additions := []string{deployment.DeployName}
		dispatch.Register(func() *UpsertResult {
			return self.AddMessageDeployments(&Message{UserId: deployment.UserId, Name: mName}, additions, nil)
		})
	} else {

		// Otherwise, create a new message from the provided Message
		deployment.Message.Deployments = []string{deployment.DeployName}
		deployment.Message.UserId = deployment.UserId
		dispatch.Register(func() *UpsertResult {
			return self.CreateMessage(deployment.Message, nil)
		})
	}
	// take list of beacon names, write the new deployname to them all
	bkns := make([]*Beacon, 0, len(deployment.BeaconNames))
	for _, bName := range deployment.BeaconNames {
		bkn := Beacon{
			UserId:     deployment.UserId,
			DeployName: deployment.DeployName,
			Name:       bName,
		}
		bkns = append(bkns, &bkn)
	}

	dispatch.Register(func() *UpsertResult {
		return self.UpdateBeacons(bkns)
	})

	// update metadata
	deploymentMeta := Deployment{
		UserId:     deployment.UserId,
		DeployName: deployment.DeployName,
		// message could be provided as a reference (MessageName) or as a new Message object
		MessageName: deployment.Message.Name,
	}

	dispatch.Register(func() *UpsertResult {
		return self.PostDeploymentMetadata(&deploymentMeta, nil)
	})

	for i := uint32(0); i < dispatch.Ct; i++ {
		res := <-dispatch.Ch

		if res.Err != nil {
			return res
		}
	}

	res := UpsertResult{
		Err: nil,
		// return nil batch b/c theres no collective batch
		Batch: nil,
	}

	return &res
}

// FetchDeploymentBeacons uses the deployments materialized view to gather a list of beacons associated with a deployment.
func (self *CassClient) FetchDeploymentBeacons(dep *Deployment) ([]*Beacon, error) {
	resRows := make([]*Beacon, 0)
	template := `SELECT user_id, deploy_name, name FROM beacon_deployments WHERE user_id = ? AND deploy_name = ? LIMIT ?`
	args := []interface{}{
		dep.UserId,
		dep.DeployName,
		DefaultLimit,
	}
	iter := self.Sess.Query(template, args...).Iter()
	shell := map[string]interface{}{}
	for iter.MapScan(shell) {
		id := shell["user_id"].(gocql.UUID)
		resRows = append(resRows, &Beacon{
			UserId:     &id,
			DeployName: shell["deploy_name"].(string),
			Name:       shell["name"].([]uint8),
		})

		// since shell is used in each iteration, we must clear it.
		shell = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return resRows, nil

}

// FetchDeployment will fetch & merge both the deployment metadata & the beacons belonging to it
func (self *CassClient) FetchDeployment(dep *Deployment) (*Deployment, error) {
	res := &Deployment{
		UserId:     dep.UserId,
		DeployName: dep.DeployName,
	}

	errCh := make(chan error, 2)
	metaCh := make(chan *Deployment)
	bknsCh := make(chan []*Beacon)

	// Fetch metadata
	go func(ch chan<- *Deployment, errCh chan<- error) {
		meta, metaErr := self.FetchDeploymentMetadata(dep.UserId, dep.DeployName)
		if metaErr != nil {
			errCh <- metaErr
			return
		}
		ch <- meta
		return

	}(metaCh, errCh)
	// Fetch beacons & merge
	go func(ch chan<- []*Beacon, errCh chan<- error) {
		bkns, err := self.FetchDeploymentBeacons(dep)
		if err != nil {
			errCh <- err
			return
		}
		ch <- bkns
		return
	}(bknsCh, errCh)

	for i := 0; i < 2; i++ {
		select {
		case meta := <-metaCh:
			res.MessageName = meta.MessageName
		case bkns := <-bknsCh:
			res.BeaconNames = mapBeaconNames(bkns)
		// If we pull an error, return through
		case err := <-errCh:
			return nil, err
		}
	}
	return res, nil
}

// Helpers

//dispatcher is a private struct which will receive commands & execute them in goroutines. You may then await the channel for responses.
// It encloses the logic for maintaining/incrementing counts
type dispatcher struct {
	Ch chan *UpsertResult
	Ct uint32
}

func (self *dispatcher) Register(fn func() *UpsertResult) {
	self.Ct += 1
	go func() {
		self.Ch <- fn()
	}()
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		Ch: make(chan *UpsertResult),
		Ct: 0,
	}
}

func mapBeaconNames(bkns []*Beacon) [][]byte {
	res := make([][]byte, 0, len(bkns))

	for _, bkn := range bkns {
		res = append(res, bkn.Name)
	}
	return res
}

func MapBytesToHex(col [][]byte) []string {
	names := make([]string, len(col))
	for i, name := range col {
		names[i] = hex.EncodeToString(name)
	}
	return names
}
