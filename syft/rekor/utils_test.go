package rekor

import (
	"fmt"
	"os"
	"testing"

	"github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/stretchr/testify/assert"
)

func Test_verifyCert(t *testing.T) {
	tests := []struct {
		name     string
		certFile string
	}{
		{
			name:     "self signed cert",
			certFile: "test-fixtures/test-certs/self-signed-cert.pem",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			certFile, err := os.Open(test.certFile)
			assert.NoError(t, err, "reading test data")
			certs, err := cryptoutils.LoadCertificatesFromPEM(certFile)
			assert.NoError(t, err, "parsing certificate")
			cert := certs[0]
			if cert == nil {
				assert.Fail(t, "reading test data")
			}

			rekorClient := &client.Rekor{}

			err = verifyCert(rekorClient, cert)
			assert.Error(t, err)
		})
	}
}

func Test_parseAndValidateAttestation(t *testing.T) {
	tests := []struct {
		name           string
		inputAttFile   string
		expectedOutput *InTotoAttestation
		expectedErr    string
		expectedLog    string
	}{
		{
			name:         "subject field is nil",
			inputAttFile: "test-fixtures/attestations/attestation-1.json",
			expectedErr:  "subject of attestation found on rekor is nil",
		},
		{
			name:         "invalid attestation json",
			inputAttFile: "test-fixtures/attestations/attestation-2.json",
			expectedErr:  "error unmarshaling attestation to inTotoAttestation type",
		},
		{
			name:         "multiple subjects",
			inputAttFile: "test-fixtures/attestations/attestation-3.json",
			expectedErr:  "multiple subjects",
		},
		{
			name:         "no predicate",
			inputAttFile: "test-fixtures/attestations/attestation-4.json",
			expectedErr:  "attestation predicate found on rekor does not contain any sboms",
		},
		{
			name:         "invalid pred type and no predicate",
			inputAttFile: "test-fixtures/attestations/attestation-5.json",
			expectedErr:  fmt.Sprintf("the attestation predicate type (foobar pred type) is not the accepted type (%v)", GoogleSbomPredicateType),
		},
		{
			name:         "no subject digest",
			inputAttFile: "test-fixtures/attestations/attestation-6.json",
			expectedErr:  "attestation subject does not contain a sha256",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			bytes, err := os.ReadFile(test.inputAttFile)
			if err != nil {
				assert.FailNow(t, "error reading test data")
			}

			logEntryAnon := &models.LogEntryAnon{
				Attestation: &models.LogEntryAnonAttestation{Data: bytes},
			}

			output, err := parseAndValidateAttestation(logEntryAnon)
			assert.Equal(t, test.expectedOutput, output)
			if test.expectedErr != "" {
				assert.ErrorContains(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// do validation of hash in subject
