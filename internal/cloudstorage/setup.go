// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cloudstorage

import (
	"context"
	"encoding/base64"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"

	"github.com/aws/aws-sdk-go/aws"
	awscreds "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/google/go-replayers/httpreplay"
	hrgoog "github.com/google/go-replayers/httpreplay/google"
	"gocloud.dev/gcp"
	"google.golang.org/api/option"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

func NewTestURLMux(t *testing.T) *blob.URLMux {
	ctx := context.Background()
	mux := new(blob.URLMux)

	// Prepare S3 client
	awsErr, awsSession, awsDone := NewAWSSession(t, "us-west-1")

	if nil == awsErr {
		s3UrlOpener := new(s3blob.URLOpener)
		s3UrlOpener.ConfigProvider = awsSession

		mux.RegisterBucket(s3blob.Scheme, s3UrlOpener)
	}

	// Prepare GCS client
	gcpError, gcpClient, gcpDone := NewGCPClient(ctx, t)

	if nil == gcpError {
		gcsUrlOpener := new(gcsblob.URLOpener)
		gcsUrlOpener.Client = gcpClient

		mux.RegisterBucket(gcsblob.Scheme, gcsUrlOpener)
	}

	// Prepare Azure client
	accountName := azureblob.AccountName("oathkeepertestbucket")

	var key azureblob.AccountKey
	if *Record {
		name, err := azureblob.DefaultAccountName()
		if err != nil {
			t.Fatal(err)
		}
		if name != accountName {
			t.Fatalf("Please update the accountName constant to match your settings file so future records work (%q vs %q)", name, accountName)
		}
		key, err = azureblob.DefaultAccountKey()
		if err != nil {
			t.Fatal(err)
		}
	} else {
		// In replay mode, we use fake credentials.
		key = azureblob.AccountKey(base64.StdEncoding.EncodeToString([]byte("FAKECREDS")))
	}

	credential, err := azureblob.NewCredential(accountName, key)
	if err != nil {
		require.NoError(t, err)
	}

	azureError, azureClient, azureDone := NewAzureTestPipeline(t, "blob", credential)
	if nil == azureError {
		azureUrlOpener := new(azureblob.URLOpener)
		azureUrlOpener.Pipeline = azureClient
		azureUrlOpener.AccountName = accountName

		mux.RegisterBucket(azureblob.Scheme, azureUrlOpener)
	}

	t.Cleanup(func() {
		if nil != awsDone {
			awsDone()
		}
		if nil != gcpDone {
			gcpDone()
		}
		if nil != azureDone {
			azureDone()
		}
	})

	return mux
}

func NewURLMux() *blob.URLMux {
	return blob.DefaultURLMux()
}

// Record is true iff the tests are being run in "record" mode.
var Record = flag.Bool("record", false, "whether to run tests against cloud resources and record the interactions")

func awsSession(region string, client *http.Client) (*session.Session, error) {
	// Provide fake creds if running in replay mode.
	var creds *awscreds.Credentials
	if !*Record {
		creds = awscreds.NewStaticCredentials("FAKE_ID", "FAKE_SECRET", "FAKE_TOKEN")
	}

	return session.NewSession(&aws.Config{
		HTTPClient:  client,
		Region:      aws.String(region),
		Credentials: creds,
		MaxRetries:  aws.Int(0),
	})
}

// NewRecordReplayClient creates a new http.Client for tests. This client's
// activity is being either recorded to files (when *Record is set) or replayed
// from files. rf is a modifier function that will be invoked with the address
// of the httpreplay.Recorder object used to obtain the client; this function
// can mutate the recorder to add service-specific header filters, for example.
// An initState is returned for tests that need a state to have deterministic
// results, for example, a seed to generate random sequences.
func NewRecordReplayClient(t *testing.T, suffix string, rf func(r *httpreplay.Recorder)) (e error, c *http.Client, cleanup func(), initState int64) { //nolint:staticcheck // legacy signature used by tests
	httpreplay.DebugHeaders()
	path := filepath.Join("testdata", t.Name()+"-"+suffix+".replay")
	if *Record {
		t.Logf("Recording into golden file %s", path)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		state := time.Now()
		b, _ := state.MarshalBinary()
		rec, err := httpreplay.NewRecorder(path, b)
		if err != nil {
			t.Fatal(err)
		}
		rf(rec)
		cleanup = func() {
			if err := rec.Close(); err != nil {
				t.Fatal(err)
			}
		}

		return nil, rec.Client(), cleanup, state.UnixNano()
	}
	t.Logf("Replaying from golden file %s", path)
	rep, err := httpreplay.NewReplayer(path)
	if err != nil {
		return err, nil, nil, 0
	}
	recState := new(time.Time)
	if err := recState.UnmarshalBinary(rep.Initial()); err != nil {
		return err, nil, nil, 0
	}

	return nil, rep.Client(), func() { _ = rep.Close() }, recState.UnixNano()
}

// NewAWSSession creates a new session for testing against AWS.
// If the test is in --record mode, the test will call out to AWS, and the
// results are recorded in a replay file.
// Otherwise, the session reads a replay file and runs the test as a replay,
// which never makes an outgoing HTTP call and uses fake credentials.
// An initState is returned for tests that need a state to have deterministic
// results, for example, a seed to generate random sequences.
func NewAWSSession(t *testing.T, region string) (e error, sess *session.Session, cleanup func()) { //nolint:staticcheck // legacy signature used by tests
	err, client, cleanup, _ := NewRecordReplayClient(t, "s3", func(r *httpreplay.Recorder) {
		r.RemoveQueryParams("X-Amz-Credential", "X-Amz-Signature", "X-Amz-Security-Token")
		r.RemoveRequestHeaders("Authorization", "Duration", "X-Amz-Security-Token")
		r.ClearHeaders("X-Amz-Date")
		r.ClearQueryParams("X-Amz-Date")
		r.ClearHeaders("User-Agent") // AWS includes the Go version
	})
	if err != nil {
		return err, nil, nil
	}
	sess, err = awsSession(region, client)
	if err != nil {
		return err, nil, nil
	}
	return nil, sess, cleanup
}

// NewGCPClient creates a new HTTPClient for testing against GCP.
// If the test is in --record mode, the client will call out to GCP, and the
// results are recorded in a replay file.
// Otherwise, the session reads a replay file and runs the test as a replay,
// which never makes an outgoing HTTP call and uses fake credentials.
func NewGCPClient(ctx context.Context, t *testing.T) (e error, client *gcp.HTTPClient, done func()) { //nolint:staticcheck // legacy signature used by tests
	err, c, cleanup, _ := NewRecordReplayClient(t, "gs", func(r *httpreplay.Recorder) {
		r.ClearQueryParams("Expires")
		r.ClearQueryParams("Signature")
		r.ClearHeaders("Expires")
		r.ClearHeaders("Signature")
	})
	if err != nil {
		return err, nil, nil
	}

	if *Record {
		creds, err := gcp.DefaultCredentials(ctx)
		if err != nil {
			return err, nil, nil
		}
		c, err = hrgoog.RecordClient(ctx, c, option.WithTokenSource(gcp.CredentialsTokenSource(creds)))
		if err != nil {
			return err, nil, nil
		}
	}
	return nil, &gcp.HTTPClient{Client: *c}, cleanup
}

// contentTypeInjectPolicy and contentTypeInjector are somewhat of a hack to
// overcome an impedance mismatch between the Azure pipeline library and
// httpreplay - the tool we use to record/replay HTTP traffic for tests.
// azure-pipeline-go does not set the Content-Type header in its requests,
// setting X-Ms-Blob-Content-Type instead; however, httpreplay expects
// Content-Type to be non-empty in some cases. This injector makes sure that
// the content type is copied into the right header when that is originally
// empty. It's only used for testing.
type contentTypeInjectPolicy struct {
	node pipeline.Policy
}

func (p *contentTypeInjectPolicy) Do(ctx context.Context, request pipeline.Request) (pipeline.Response, error) {
	if len(request.Header.Get("Content-Type")) == 0 {
		cType := request.Header.Get("X-Ms-Blob-Content-Type")
		request.Header.Set("Content-Type", cType)
	}
	response, err := p.node.Do(ctx, request)
	return response, err
}

type contentTypeInjector struct {
}

func (f contentTypeInjector) New(node pipeline.Policy, opts *pipeline.PolicyOptions) pipeline.Policy {
	return &contentTypeInjectPolicy{node: node}
}

// NewAzureTestPipeline creates a new connection for testing against Azure Blob.
func NewAzureTestPipeline(t *testing.T, api string, credential azblob.Credential) (error, pipeline.Pipeline, func()) { //nolint:staticcheck // legacy signature used by tests
	err, client, done, _ := NewRecordReplayClient(t, "azblob", func(r *httpreplay.Recorder) {
		r.RemoveQueryParams("se", "sig")
		r.RemoveQueryParams("X-Ms-Date")
		r.ClearHeaders("X-Ms-Date")
		r.ClearHeaders("User-Agent") // includes the full Go version
	})
	if err != nil {
		return err, nil, nil
	}

	f := []pipeline.Factory{
		// Sets User-Agent for recorder.
		azblob.NewTelemetryPolicyFactory(azblob.TelemetryOptions{
			Value: AzureUserAgentPrefix(api),
		}),
		contentTypeInjector{},
		credential,
		pipeline.MethodFactoryMarker(),
	}
	// Create a pipeline that uses client to make requests.
	p := pipeline.NewPipeline(f, pipeline.Options{
		HTTPSender: pipeline.FactoryFunc(func(next pipeline.Policy, po *pipeline.PolicyOptions) pipeline.PolicyFunc {
			return func(ctx context.Context, request pipeline.Request) (pipeline.Response, error) {
				r, err := client.Do(request.WithContext(ctx))
				if err != nil {
					err = pipeline.NewError(err, "HTTP request failed")
				}
				return pipeline.NewHTTPResponse(r), err
			}
		}),
	})

	return nil, p, done
}
