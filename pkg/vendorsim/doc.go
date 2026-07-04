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

// Package vendorsim provides shared building blocks for the external-vendor
// simulators under services/simulators/* (payment rails, card processor,
// KYC vendor, FX rates): signed webhook delivery with retries, consumer-side
// verification and de-duplication, deterministic "magic value" outcomes, and
// chaos knobs for delay/duplicate/reorder testing.
//
// A simulator schedules outbound webhooks with Dispatcher; a domain service
// consuming them wraps its webhook handler with VerifyWebhook. Both sides
// agree on the magic-value conventions documented in magicvalues.go.
//
// See docs/vendor-simulators-plan.md for the overall design this supports.
package vendorsim
