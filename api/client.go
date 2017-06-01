package api

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// Instantiate a client with credentials bound
func JWTConfigFromJSON(fPath, scope string) *http.Client {
	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	// Navigate to your project, then see the "Credentials" page
	// under "APIs & Auth".
	// To create a service account client, click "Create new Client ID",
	// select "Service Account", and click "Create Client ID". A JSON
	// key file will then be downloaded to your computer.
	data, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(data, scope)
	if err != nil {
		log.Fatal(err)
	}
	// Initiate an http.Client. The following GET request will be
	// authorized and authenticated on the behalf of
	// your service account.
	client := conf.Client(oauth2.NoContext)

	return client
}

type BeaconClient struct {
	Svc *proximitybeacon.Service
}

func NewBeaconClient(client *http.Client) (*BeaconClient, error) {
	svc, err := proximitybeacon.New(client)
	if err != nil {
		return nil, err
	}
	return &BeaconClient{svc}, nil

}

func (c *BeaconClient) GetOwnedBeaconNames() (*proximitybeacon.ListBeaconsResponse, error) {
	return c.Svc.Beacons.List().Do()
}

func (c *BeaconClient) GetBeaconById(name string) (*proximitybeacon.Beacon, error) {
	return c.Svc.Beacons.Get(name).Do()
}

func (c *BeaconClient) GetBeaconsByNames(bNames []string) []*proximitybeacon.Beacon {
	var wg sync.WaitGroup
	length := len(bNames)
	wg.Add(length)
	results := make([]*proximitybeacon.Beacon, length)
	// process concurently in goroutines
	for i, name := range bNames {
		go func(i int) {
			beacon, _ := c.GetBeaconById(name)
			// TBD: add error handling, possibly pushing errs to its own slice
			results[i] = beacon
			wg.Done()
		}(i)
	}

	wg.Wait()

	return results
}
