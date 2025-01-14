// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resourceusage

import (
	"context"
	"time"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"

	"github.com/gravitational/teleport/api/types"
	apievents "github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/teleport/lib/events"
)

// GetAccessRequestMonthlyUsage returns the number of access requests that have been created this month.
func GetAccessRequestMonthlyUsage(ctx context.Context, alog events.AuditLogger, now time.Time) (int, error) {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	created := make(map[string]struct{})

	var results []apievents.AuditEvent
	var startKey string
	var err error
	for {
		results, startKey, err = alog.SearchEvents(ctx, events.SearchEventsRequest{
			From:       monthStart,
			To:         now,
			Order:      types.EventOrderAscending,
			EventTypes: []string{events.AccessRequestCreateEvent},
			StartKey:   startKey,
		})
		if err != nil {
			return 0, trace.Wrap(err)
		}
		for _, ev := range results {
			ev, ok := ev.(*apievents.AccessRequestCreate)
			if !ok {
				return 0, trace.BadParameter("expected *AccessRequestCreate, but got %T", ev)
			}
			id := ev.RequestID
			switch ev.GetType() {
			case events.AccessRequestCreateEvent:
				created[id] = struct{}{}
			default:
				log.Warnf("Expected event type %q, got %q", events.AccessRequestCreateEvent, ev.GetType())
			}
		}
		if startKey == "" {
			break
		}
	}

	return len(created), nil
}
