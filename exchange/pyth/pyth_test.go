// Copyright (C) 2025 Quantive
//
// SPDX-License-Identifier: MIT OR AGPL-3.0-or-later
//
// This file is part of the Decision Engine project.
// You may choose to use this file under the terms of either
// the MIT License or the GNU Affero General Public License v3.0 or later.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the LICENSE files for more details.

package pyth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
)

const testFeedID = "0xabc123"

func testPair() model.QuotesPair {
	return model.QuotesPair{
		ExchangeID: model.PYTH,
		Base:       currency.CurrencySymbol("BTC"),
		Quote:      currency.CurrencySymbol("USD"),
	}
}

func TestGetPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/updates/price/latest" {
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
		if r.URL.Query().Get("parsed") != "true" || r.URL.Query().Get("encoding") != "base64" {
			t.Errorf("missing parsed price query parameters: %s", r.URL.RawQuery)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("Authorization header = %q, want bearer token", got)
		}
		if got := r.URL.Query()["ids[]"]; len(got) != 1 || got[0] != testFeedID {
			t.Fatalf("unexpected feed ID query: %v", got)
		}
		_, _ = w.Write([]byte(`{"parsed":[{"id":"0xabc123","price":{"price":"123456789","expo":-6,"publish_time":1700000000}}]}`))
	}))
	defer server.Close()

	client := NewSingleClient(PythConfig{
		FeedIDs:        map[string]string{"btc/usd": testFeedID},
		HermesEndpoint: server.URL,
		AccessToken:    "test-access-token",
	})
	point, err := client.GetPrice(context.Background(), testPair())
	if err != nil {
		t.Fatal(err)
	}
	if want := decimal.RequireFromString("123.456789"); !point.NewPrice.Equal(want) {
		t.Errorf("price = %s, want %s", point.NewPrice, want)
	}
	if want := time.Unix(1700000000, 0); !point.UpdatedAt.Equal(want) {
		t.Errorf("updated at = %s, want %s", point.UpdatedAt, want)
	}
}

func TestGetPriceRequiresConfiguredFeed(t *testing.T) {
	client := NewSingleClient(PythConfig{})
	_, err := client.GetPrice(context.Background(), testPair())
	if !errors.Is(err, errFeedNotFound) {
		t.Fatalf("GetPrice error = %v, want %v", err, errFeedNotFound)
	}
}

func TestUnsupportedProviderOperations(t *testing.T) {
	client := New(PythConfig{})
	pair := testPair()

	if _, err := client.GetKlines(context.Background(), pair, "1m", 10); !errors.Is(err, ErrUnsupported) {
		t.Errorf("GetKlines error = %v, want %v", err, ErrUnsupported)
	}
	if _, err := client.GetOrderBook(context.Background(), pair, 10); !errors.Is(err, ErrUnsupported) {
		t.Errorf("GetOrderBook error = %v, want %v", err, ErrUnsupported)
	}
	if _, err := client.PlaceOrder(context.Background(), model.OrderRequest{}); !errors.Is(err, ErrUnsupported) {
		t.Errorf("PlaceOrder error = %v, want %v", err, ErrUnsupported)
	}
	if _, err := client.GetOrder(context.Background(), "", ""); !errors.Is(err, ErrUnsupported) {
		t.Errorf("GetOrder error = %v, want %v", err, ErrUnsupported)
	}
	if err := client.CancelOrder(context.Background(), "", ""); !errors.Is(err, ErrUnsupported) {
		t.Errorf("CancelOrder error = %v, want %v", err, ErrUnsupported)
	}
	if _, err := client.GetAssetBalance(context.Background(), "BTC"); !errors.Is(err, ErrUnsupported) {
		t.Errorf("GetAssetBalance error = %v, want %v", err, ErrUnsupported)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStreamSubscribeAndDispatchPriceUpdate(t *testing.T) {
	connected := make(chan struct{})
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("ACCESS_TOKEN"); got != "test-access-token" {
			t.Errorf("ACCESS_TOKEN query = %q, want test access token", got)
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var subscription struct {
			Type    string   `json:"type"`
			IDs     []string `json:"ids"`
			Verbose bool     `json:"verbose"`
			Binary  bool     `json:"binary"`
		}
		if json.Unmarshal(message, &subscription) != nil ||
			subscription.Type != "subscribe" ||
			len(subscription.IDs) != 1 ||
			subscription.IDs[0] != testFeedID ||
			!subscription.Verbose ||
			!subscription.Binary {
			return
		}
		close(connected)
		_ = conn.WriteJSON(map[string]interface{}{
			"type": "price_update",
			"price_feed": map[string]interface{}{
				"id": testFeedID,
				"price": map[string]interface{}{
					"price":        "123456789",
					"expo":         -6,
					"publish_time": int64(1700000000),
				},
			},
		})
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	client := NewStreamClient(PythConfig{
		FeedIDs:          map[string]string{"BTC/USD": testFeedID},
		HermesWSEndpoint: "ws" + strings.TrimPrefix(server.URL, "http"),
		AccessToken:      "test-access-token",
	})
	defer client.Close()
	if err := client.SubscribeStream(testPair(), nil); err != nil {
		t.Fatal(err)
	}
	priceChan, _, _ := client.ReceiveStream()
	ctx, cancel := context.WithCancel(context.Background())
	dispatchDone := make(chan error, 1)
	go func() {
		dispatchDone <- client.Dispatch(ctx)
	}()

	select {
	case <-connected:
	case <-time.After(2 * time.Second):
		t.Fatal("websocket subscription was not received")
	}
	select {
	case point := <-priceChan:
		if want := decimal.RequireFromString("123.456789"); !point.NewPrice.Equal(want) {
			t.Errorf("stream price = %s, want %s", point.NewPrice, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for price update")
	}
	cancel()
	select {
	case err := <-dispatchDone:
		if err != nil {
			t.Fatalf("Dispatch returned error after cancellation: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Dispatch did not stop after cancellation")
	}
}
