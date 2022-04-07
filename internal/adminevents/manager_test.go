// Copyright © 2022 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adminevents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hyperledger/firefly/internal/config"
	"github.com/hyperledger/firefly/internal/config/wsconfig"
	"github.com/hyperledger/firefly/internal/restclient"
	"github.com/hyperledger/firefly/pkg/fftypes"
	"github.com/hyperledger/firefly/pkg/wsclient"
	"github.com/stretchr/testify/assert"
)

func newTestAdminEventsManager(t *testing.T) (ae *adminEventManager, wsc wsclient.WSClient, cancel func()) {
	config.Reset()

	ae = NewAdminEventManager(context.Background()).(*adminEventManager)
	svr := httptest.NewServer(http.HandlerFunc(ae.ServeHTTPWebSocketListener))

	clientPrefix := config.NewPluginConfig("ut.wsclient")
	wsconfig.InitPrefix(clientPrefix)
	clientPrefix.Set(restclient.HTTPConfigURL, fmt.Sprintf("http://%s", svr.Listener.Addr()))
	wsConfig := wsconfig.GenerateConfigFromPrefix(clientPrefix)

	wsc, err := wsclient.New(ae.ctx, wsConfig, nil, nil)
	assert.NoError(t, err)
	err = wsc.Connect()
	assert.NoError(t, err)

	return ae, wsc, func() {
		ae.cancelCtx()
		wsc.Close()
		ae.WaitStop()
		svr.Close()
	}
}

func toJSON(t *testing.T, obj interface{}) []byte {
	b, err := json.Marshal(obj)
	assert.NoError(t, err)
	return b
}

func TestAdminEventsE2E(t *testing.T) {
	ae, wsc, cancel := newTestAdminEventsManager(t)

	events := make(chan *fftypes.ChangeEvent)
	go func() {
		for msgBytes := range wsc.Receive() {
			var event fftypes.ChangeEvent
			err := json.Unmarshal(msgBytes, &event)
			assert.NoError(t, err)
			events <- &event
		}
	}()

	// Send some garbage first, to be discarded
	wsc.Send(ae.ctx, toJSON(t, map[string]string{"wrong": "data"}))
	// Then send the actual command to start with a filter
	wsc.Send(ae.ctx, toJSON(t, &fftypes.WSChangeEventCommand{
		Type:        fftypes.WSChangeEventCommandTypeStart,
		Collections: []string{"collection1"},
		Filter: fftypes.ChangeEventFilter{
			Types:      []fftypes.ChangeEventType{fftypes.ChangeEventTypeCreated},
			Namespaces: []string{"ns1"},
		},
	}))
	for len(ae.dirtyReadList) == 0 || ae.dirtyReadList[0].collections == nil {
		time.Sleep(1 * time.Microsecond)
	}

	ignoreDueToWrongCollection := &fftypes.ChangeEvent{
		Collection: "collection2",
		Type:       fftypes.ChangeEventTypeCreated,
		Namespace:  "ns1",
	}
	ae.Dispatch(ignoreDueToWrongCollection)
	ignoreDueToWrongType := &fftypes.ChangeEvent{
		Collection: "collection1",
		Type:       fftypes.ChangeEventTypeDeleted,
		Namespace:  "ns1",
	}
	ae.Dispatch(ignoreDueToWrongType)
	ignoreDueToWrongNamespace := &fftypes.ChangeEvent{
		Collection: "collection1",
		Type:       fftypes.ChangeEventTypeCreated,
		Namespace:  "ns2",
	}
	ae.Dispatch(ignoreDueToWrongNamespace)
	match := &fftypes.ChangeEvent{
		Collection: "collection1",
		Type:       fftypes.ChangeEventTypeCreated,
		Namespace:  "ns1",
	}
	ae.Dispatch(match)

	assert.Equal(t, match, <-events)

	defer cancel()

}

func TestBadUpgrade(t *testing.T) {
	config.Reset()

	ae := NewAdminEventManager(context.Background()).(*adminEventManager)
	svr := httptest.NewServer(http.HandlerFunc(ae.ServeHTTPWebSocketListener))
	defer svr.Close()

	res, err := http.Post(fmt.Sprintf("http://%s", svr.Listener.Addr()), "application/json", bytes.NewReader([]byte("{}")))
	assert.NoError(t, err)
	assert.True(t, res.StatusCode >= 300)

}