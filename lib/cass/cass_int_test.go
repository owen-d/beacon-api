package cass

import (
	"encoding/hex"
	"github.com/gocql/gocql"
	"log"
	"testing"
)

const (
	prepopId    = "6ba7b810-9dad-11d1-80b4-00c04fd430c9"
	prepopEmail = "int-test.email@gmail.com"
	prepopMName = "first_msg"
	prepopDName = "dep1"
)

var (
	prepopBName, _ = hex.DecodeString("00000000000000000000000000000000")
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

func TestFulfillsInterface(t *testing.T) {
	var _ Client = createLocalhostClient("bkn")
}

func TestCreateBeacons(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()
	uuid, _ := gocql.ParseUUID(prepopId)

	t.Run("non-batch", func(t *testing.T) {
		bkns := []*Beacon{
			&Beacon{
				UserId: &uuid,
				Name:   []byte("testcreate-1"),
			},
			&Beacon{
				UserId: &uuid,
				Name:   []byte("testcreate-2"),
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
				Name:   []byte("testcreate-b-1"),
			},
			&Beacon{
				UserId: &uuid,
				Name:   []byte("testcreate-b-2"),
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
		MsgUrl:     "http://fake.url",
	}

	res := client.UpdateBeacons([]*Beacon{&bkn})
	if res.Err != nil {
		t.Error("failed to create beacons: %v", res.Err)
	}
}

func TestRemoveBeaconsDeployments(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)
	bkn := Beacon{
		UserId: &uuid,
		Name:   prepopBName,
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

	found, err := client.FetchBeacon(&bkn)
	if err != nil || found == nil {
		t.Error("failed to match beacon:", err, ", ", found)
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

func TestUpdateMessage(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)
	msg := Message{
		UserId: &uuid,
		Name:   prepopMName,
		Title:  "newtitle",
		Url:    "http://newurl.sharecro.ws",
	}

	t.Run("update", func(t *testing.T) {
		res := client.UpdateMessage(&msg, nil)
		if res.Err != nil {
			t.Error("failed to update msg:", res.Err)
		}
	})
	t.Run("update-batch", func(t *testing.T) {
		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.UpdateMessage(&msg, batch)

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

	fetched, err := client.FetchMessage(&msg)

	if err != nil || fetched == nil {
		t.Error("failed to fetch msg:", err)
		return
	}

}

func TestFetchMessages(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	msg := Message{
		UserId: &uuid,
	}

	fetched, err := client.FetchMessages(msg.UserId, 5)

	if err != nil || len(fetched) == 0 {
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

		_, err := client.FetchDeploymentsMetadata(dep.UserId)

		if err != nil {
			t.Error(err)
		}
	})
}

func TestFetchDeploymentMetadata(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)
	dep := &Deployment{
		UserId:     &uuid,
		DeployName: prepopDName,
	}
	t.Run("single", func(t *testing.T) {
		fetched, err := client.FetchDeploymentMetadata(dep.UserId, dep.DeployName)
		if err != nil || fetched == nil {
			t.Error(err)
		}
	})
	t.Run("multi", func(t *testing.T) {
		fetched, err := client.FetchDeploymentsMetadata(dep.UserId)
		if err != nil || len(fetched) == 0 {
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
			BeaconNames: [][]byte{prepopBName},
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
			BeaconNames: [][]byte{prepopBName},
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
			BeaconNames: [][]byte{prepopBName},
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

func TestFetchDeploymentBeacons(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	dep := Deployment{
		UserId:     &uuid,
		DeployName: prepopDName,
	}

	fetched, err := client.FetchDeploymentBeacons(&dep)

	if err != nil || len(fetched) == 0 {
		t.Error(err)
	}

}

func TestFetchDeployment(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	uuid, _ := gocql.ParseUUID(prepopId)

	dep := Deployment{
		UserId:     &uuid,
		DeployName: prepopDName,
	}

	fetched, err := client.FetchDeployment(&dep)

	if err != nil || fetched == nil {
		t.Error(err)
	}
}
