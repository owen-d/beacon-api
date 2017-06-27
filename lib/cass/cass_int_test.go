package cass

import (
	"github.com/gocql/gocql"
	"log"
	"testing"
)

const (
	prepopId    = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	prepopEmail = "test.email@gmail.com"
	prepopBName = "a1"
	prepopMName = "first_msg"
)

// need to be able to call subtest on cmd.

func createLocalhostClient(keyspace string) *CassClient {
	cluster := gocql.NewCluster("localhost")
	client, err := Connect(cluster, keyspace)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func testBatch(t *testing.T, res *UpsertResult, client *CassClient, batch *gocql.Batch) {
	if res.Err != nil {
		t.Error("failed to create batch", res.Err)
	}
	// check attempt ct
	if res.Batch.Attempts() != 0 {
		t.Error("batch preemptively executed")
	}

	// execute batch
	err := client.Sess.ExecuteBatch(res.Batch)
	if err != nil {
		t.Error("failed batch execution", err)
	}
}

func TestCreateUser(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	t.Run("Individual", func(t *testing.T) {
		uuid, _ := gocql.RandomUUID()
		newUser := User{
			Id:    &uuid,
			Email: "newEmail@provider.com",
		}
		res := client.CreateUser(&newUser, nil)
		if res.Err != nil {
			t.Fail()
		}
	})
	t.Run("Batch", func(t *testing.T) {
		uuid, _ := gocql.RandomUUID()
		newUser := User{
			Id:    &uuid,
			Email: "newEmail-batch@provider.com",
		}
		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.CreateUser(&newUser, batch)

		// check size
		if res.Batch.Size() != 1 {
			t.Error("batch has incorerct # of statements:")
		}

		testBatch(t, res, client, batch)

	})
}

func TestFetchUser(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	t.Run("uuid", func(t *testing.T) {
		// propulated user id
		uuid, parseErr := gocql.ParseUUID(prepopId)

		if parseErr != nil {
			t.Error("failed parsing uuid")
			return
		}

		user := User{
			Id: &uuid,
		}

		foundUser, fetchErr := client.FetchUser(&user)

		if fetchErr != nil {
			t.Error("failed fetching user", fetchErr)
			return
		}

		if foundUser == nil {
			t.Error("failed to match user")
			return
		}

	})

	t.Run("email", func(t *testing.T) {
		user := User{
			Email: prepopEmail,
		}

		foundUser, fetchErr := client.FetchUser(&user)

		if fetchErr != nil {
			t.Error("failed fetching user", fetchErr)
			return
		}

		if foundUser == nil {
			t.Error("failed to match user")
			return
		}
	})
}

func TestCreateBeacons(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()
	uuid, _ := gocql.ParseUUID(prepopId)

	t.Run("non-batch", func(t *testing.T) {
		bkns := []*Beacon{
			&Beacon{
				UserId: &uuid,
				Name:   "testcreate-1",
			},
			&Beacon{
				UserId: &uuid,
				Name:   "testcreate-2",
			},
		}

		res := client.CreateBeacons(bkns, nil)
		if res.Err != nil {
			t.Error("failed to create beacons: %v", res.Err)
		}

	})

	t.Run("batch", func(t *testing.T) {
		bkns := []*Beacon{
			&Beacon{
				UserId: &uuid,
				Name:   "testcreate-b-1",
			},
			&Beacon{
				UserId: &uuid,
				Name:   "testcreate-b-2",
			},
		}

		batch := gocql.NewBatch(gocql.LoggedBatch)

		res := client.CreateBeacons(bkns, batch)
		testBatch(t, res, client, batch)
	})
}

func TestUpdateBeacons(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)
	bkn := Beacon{
		UserId:     &uuid,
		Name:       prepopBName,
		DeployName: "non-batch-deploy",
	}

	res := client.UpdateBeacons([]*Beacon{&bkn})
	if res.Err != nil {
		t.Error("failed to create beacons: %v", res.Err)
	}
}

func TestFetchBeacon(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	bkn := Beacon{
		UserId: &uuid,
		Name:   prepopBName,
	}

	_, err := client.FetchBeacon(&bkn)
	if err != nil {
		t.Error("failed to match beacon:", err)
	}
}

func TestCreateMessage(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	t.Run("non-batch", func(t *testing.T) {
		msg := Message{
			UserId:      &uuid,
			Name:        "non-batch-create-msg",
			Title:       "filler",
			Url:         "https://filler.com",
			Lang:        "en",
			Deployments: []string{"dep1"},
		}

		res := client.CreateMessage(&msg, nil)
		if res.Err != nil {
			t.Error("failed to create msg:", res.Err)
		}

	})

	t.Run("batch", func(t *testing.T) {
		msg := Message{
			UserId:      &uuid,
			Name:        "non-batch-create-msg",
			Title:       "filler",
			Url:         "https://filler.com",
			Lang:        "en",
			Deployments: []string{"dep1"},
		}

		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.CreateMessage(&msg, batch)

		testBatch(t, res, client, batch)
	})
}

func TestChangeMessageDeployments(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)
	msg := Message{
		UserId: &uuid,
		Name:   prepopMName,
	}

	t.Run("add", func(t *testing.T) {
		additions := []string{"add1", "add2"}
		res := client.AddMessageDeployments(&msg, additions, nil)
		if res.Err != nil {
			t.Fail()
		}
	})

	t.Run("add-batch", func(t *testing.T) {
		additions := []string{"add1-batch", "add2-batch"}
		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.AddMessageDeployments(&msg, additions, batch)

		testBatch(t, res, client, batch)
	})

	t.Run("remove", func(t *testing.T) {
		removals := []string{"remove1", "remove2"}
		res := client.RemoveMessageDeployments(&msg, removals, nil)
		if res.Err != nil {
			t.Fail()
		}
	})

	t.Run("remove-batch", func(t *testing.T) {
		removals := []string{"remove1-batch", "remove2-batch"}
		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.RemoveMessageDeployments(&msg, removals, batch)

		testBatch(t, res, client, batch)
	})
}

func TestFetchMessage(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	msg := Message{
		UserId: &uuid,
		Name:   prepopMName,
	}

	_, err := client.FetchMessage(&msg)

	if err != nil {
		t.Error("failed to fetch msg:", err)
		return
	}

}

func TestPostDeploymentMetadata(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	t.Run("individual", func(t *testing.T) {
		dep := Deployment{
			UserId:      &uuid,
			DeployName:  "dep_md_create",
			MessageName: "msg",
		}

		res := client.PostDeploymentMetadata(&dep, nil)

		if res.Err != nil {
			t.Fail()
		}
	})

	t.Run("batch", func(t *testing.T) {
		dep := Deployment{
			UserId:      &uuid,
			DeployName:  "dep_md_create",
			MessageName: "msg",
		}

		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.PostDeploymentMetadata(&dep, batch)

		testBatch(t, res, client, batch)
	})

	t.Run("fetch", func(t *testing.T) {
		dep := Deployment{
			UserId: &uuid,
		}

		_, err := client.FetchDeploymentsMetadata(&dep)

		if err != nil {
			t.Error(err)
		}
	})

}

func TestPostDeployment(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	t.Run("from-MessageName", func(t *testing.T) {
		dep := Deployment{
			UserId:      &uuid,
			DeployName:  "test-full-deployment",
			MessageName: prepopMName,
			BeaconNames: []string{prepopBName},
		}

		res := client.PostDeployment(&dep)
		if res.Err != nil {
			t.Error("failed to post deployment:", res.Err)
		}
	})

	t.Run("from-MessageName-notexists", func(t *testing.T) {
		dep := Deployment{
			UserId:      &uuid,
			DeployName:  "test-full-deployment",
			MessageName: "deploytest-message-name-invalid",
			BeaconNames: []string{prepopBName},
		}

		res := client.PostDeployment(&dep)
		if res.Err == nil {
			t.Error("false positive: should have failed with invalid message name")
		}
	})

	t.Run("from-Message", func(t *testing.T) {
		dep := Deployment{
			UserId:      &uuid,
			DeployName:  "test-full-deployment",
			BeaconNames: []string{prepopBName},
			Message: &Message{
				Name:  "deploytest-new-msg",
				Title: "hi",
				Url:   "https://google.com",
				Lang:  "en",
			},
		}

		res := client.PostDeployment(&dep)
		if res.Err != nil {
			t.Error("failed to post deployment:", res.Err)
		}
	})
}
