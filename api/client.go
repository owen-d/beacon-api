package api

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"io/ioutil"
	"log"
	"net/http"
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
	length := len(bNames)
	type Wrapper struct {
		I   int
		Ptr *proximitybeacon.Beacon
		Err error
	}
	ch := make(chan *Wrapper, length)
	results := make([]*proximitybeacon.Beacon, length)
	// process concurently in goroutines
	for i, name := range bNames {
		go func(i int) {
			beacon, err := c.GetBeaconById(name)
			ch <- &Wrapper{i, beacon, err}
		}(i)
	}

	for i := 0; i < length; i++ {
		wrapper := <-ch
		// TBD: add error handling
		results[wrapper.I] = wrapper.Ptr
	}

	return results
}
