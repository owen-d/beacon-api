package cass

import (
	"github.com/gocql/gocql"
	"log"
	"testing"
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

func TestCreateUser(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.close()

	t.Run("Individual", func(t *testing.T) {
		newUser := User{
			Id:    gocql.RandomUUID(),
			Email: "newEmail@provider.com",
		}
		res := client.CreateUser(&new, nil)
		if res.Err != nil {
			t.Fail()
		}
	})
	t.Run("Batch", func(t *testing.T) {
		newUser := User{
			Id:    gocql.RandomUUID(),
			Email: "newEmail-batch@provider.com",
		}
		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.CreateUser(&new, batch)

		if res.Err != nil {
			t.Error("failed to create batch", res.Err)
		}
		// check attempt ct
		if res.Batch.Attempts() != 0 {
			t.Error("batch preemptively executed")
		}
		// check size
		if res.Batch.Size() != 1 {
			t.Error("batch has incorerct # of statements:")
		}

		// execute batch
		err := client.Sess.ExecuteBatch(res.Batch)
		if err != nil {
			t.Error("failed batch execution", err)
		}
	})
}

func TestFetchUser(t *Testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.close()
}
