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

package outbox

import "context"

// MultiProducer fans out each event to every configured producer.
type MultiProducer struct {
	producers []Producer
}

func NewMultiProducer(producers ...Producer) MultiProducer {
	filtered := make([]Producer, 0, len(producers))
	for _, p := range producers {
		if p != nil {
			filtered = append(filtered, p)
		}
	}

	return MultiProducer{producers: filtered}
}

func (m MultiProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	for _, p := range m.producers {
		err := p.Produce(ctx, topic, key, value)
		if err != nil {
			return err
		}
	}

	return nil
}
