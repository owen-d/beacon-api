package cass

import (
	"github.com/gocql/gocql"
	"math/rand"
	"testing"
)

func randToken() []byte {
	token := make([]byte, 16)
	rand.Read(token)
	return token

}

func TestCreateUser(t *testing.T) {
	client := createLocalhostClient("bkn")
	defer client.Sess.Close()

	t.Run("Individual", func(t *testing.T) {
		newUser := User{
			Email: "newEmail@provider.com",
		}
		res := client.CreateUser(&newUser, Google, randToken(), nil)
		if res.Err != nil {
			t.Fail()
		}
	})
	t.Run("Batch", func(t *testing.T) {
		newUser := User{
			Email: "newEmail-batch@provider.com",
		}
		batch := gocql.NewBatch(gocql.LoggedBatch)
		res := client.CreateUser(&newUser, Google, randToken(), batch)

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
