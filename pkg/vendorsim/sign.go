//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vendorsim

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	// HeaderTimestamp carries the unix-seconds timestamp a delivery was signed at.
	HeaderTimestamp = "X-Vendorsim-Timestamp"
	// HeaderSignature carries the hex HMAC-SHA256 signature over the timestamp and body.
	HeaderSignature = "X-Vendorsim-Signature"
	// HeaderEventType carries the simulator-defined event type (e.g. "rails.transfer.received").
	HeaderEventType = "X-Vendorsim-Event"
	// HeaderDeliveryID carries the unique delivery ID, used by consumers for replay detection.
	HeaderDeliveryID = "X-Vendorsim-Delivery-Id"
)

var (
	ErrMissingSignature  = errors.New("vendorsim: missing signature headers")
	ErrBadTimestamp      = errors.New("vendorsim: invalid timestamp header")
	ErrSignatureExpired  = errors.New("vendorsim: signature timestamp outside allowed skew")
	ErrSignatureMismatch = errors.New("vendorsim: signature mismatch")
)

// Sign computes the signature a simulator attaches to a webhook delivery,
// over "<unix-timestamp>.<body>". Consumers verify it with VerifySignature.
func Sign(secret []byte, timestamp int64, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature checks a webhook delivery's timestamp and signature header
// values against secret, rejecting timestamps further than maxSkew from now.
func VerifySignature(secret []byte, timestampHeader, signatureHeader string, body []byte, maxSkew time.Duration) error {
	if timestampHeader == "" || signatureHeader == "" {
		return ErrMissingSignature
	}

	ts, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadTimestamp, err)
	}

	age := time.Since(time.Unix(ts, 0))
	if age < 0 {
		age = -age
	}

	if age > maxSkew {
		return ErrSignatureExpired
	}

	expected := Sign(secret, ts, body)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(signatureHeader)) != 1 {
		return ErrSignatureMismatch
	}

	return nil
}
