package beaconclient

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	googleNamespacedType = "com.google.nearby/en"
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

type Client interface {
	GetOwnedBeaconNames() (*proximitybeacon.ListBeaconsResponse, error)
	GetBeaconById(name string) (*proximitybeacon.Beacon, error)
	GetBeaconsByNames(bNames []string) []*proximitybeacon.Beacon
	GetAttachmentsForBeacon(name string) ([]*proximitybeacon.BeaconAttachment, error)
	CreateAttachment(beaconName string, attachmentData *AttachmentData) (*proximitybeacon.BeaconAttachment, error)
	BatchDeleteAttachments(beaconName string) (int64, error)
	DeclarativeAttach([][]byte, *AttachmentData) []*AttachmentResult
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
	return c.Svc.Beacons.List().Q("status:active").Do()
}

func (c *BeaconClient) GetBeaconById(name string) (*proximitybeacon.Beacon, error) {
	prefixed := "beacons/3!" + name
	return c.Svc.Beacons.Get(prefixed).Do()
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

func (c *BeaconClient) GetAttachmentsForBeacon(name string) ([]*proximitybeacon.BeaconAttachment, error) {
	prefixed := "beacons/3!" + name
	res, err := c.Svc.Beacons.Attachments.List(prefixed).NamespacedType(googleNamespacedType).Do()
	var results []*proximitybeacon.BeaconAttachment
	if err != nil {
		return results, err
	}

	return append(results, res.Attachments...), nil
}

// TBD: parameterize namespacedType
func (c *BeaconClient) CreateAttachment(beaconName string, attachmentData *AttachmentData) (*proximitybeacon.BeaconAttachment, error) {
	prefixed := "beacons/3!" + beaconName
	data := attachmentData.encode()
	newAttachment := proximitybeacon.BeaconAttachment{
		Data:           data,
		NamespacedType: googleNamespacedType,
	}

	return c.Svc.Beacons.Attachments.Create(prefixed, &newAttachment).Do()
}

// TBD: parameterize namespacedType
func (c *BeaconClient) BatchDeleteAttachments(beaconName string) (int64, error) {
	prefixed := "beacons/3!" + beaconName
	res, err := c.Svc.Beacons.Attachments.BatchDelete(prefixed).NamespacedType(googleNamespacedType).Do()
	if err != nil {
		return 0, err
	}
	return res.NumDeleted, nil
}

type AttachmentData struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (a *AttachmentData) encode() string {
	jData, _ := json.Marshal(a)
	return base64.StdEncoding.EncodeToString(jData)
}

// AttachmentResult is a wrapper type hol,ding response data from google beacon platform about attachment deletions and creations
type AttachmentResult struct {
	Name       string
	Err        error
	Attachment *proximitybeacon.BeaconAttachment `json:-`
}

func (self *BeaconClient) DeclarativeAttach(bNames [][]byte, attachment *AttachmentData) []*AttachmentResult {
	res := make([]*AttachmentResult, 0, len(bNames))

	ch := make(chan *AttachmentResult)

	// delete old attachments & apply new one
	for _, bName := range bNames {
		go func(bName []byte, ch chan<- *AttachmentResult) {
			// assign url altered url
			strName := hex.EncodeToString(bName)
			shortBknName := bName[len(bName)-6:]
			alteredAttach := &AttachmentData{
				Title: attachment.Title,
				Url:   fmt.Sprint("https://our.sharecro.ws/bkn/", hex.EncodeToString(shortBknName)),
			}

			resp := &AttachmentResult{Name: strName}

			// remove old attachments on beacon
			_, deleteErr := self.BatchDeleteAttachments(strName)
			if deleteErr != nil {
				resp.Err = deleteErr
				ch <- resp
				return
			}

			// early return for nil attachments (functions as just the delete from prev step)
			if attachment == nil {
				ch <- resp
				return
			}

			postedAttachment, postErr := self.CreateAttachment(strName, alteredAttach)

			if postErr != nil {
				resp.Err = postErr
				ch <- resp
				return
			} else {
				resp.Attachment = postedAttachment
				ch <- resp
				return
			}
		}(bName, ch)
	}

	for range bNames {
		res = append(res, <-ch)
	}

	return res
}
