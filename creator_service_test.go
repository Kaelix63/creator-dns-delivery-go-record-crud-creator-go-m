package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeliveryDefaultsToCNAMEAndUsesZone(t *testing.T) {
	tests := []struct {
		name     string
		input    DeliveryRequest
		wantType string
	}{
		{name: "delivery host defaults to cname", input: DeliveryRequest{Domain: "studio.example", Name: "download", Content: "cdn.example"}, wantType: "CNAME"},
		{name: "explicit txt remains txt", input: DeliveryRequest{Domain: "studio.example", Name: "_dmarc", Content: "v=DMARC1", Type: "TXT"}, wantType: "TXT"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v1/dns/domain/get" {
					_, _ = w.Write([]byte(`{"ok":true,"data":{"zone_id":"zone-7"},"error":null,"metadata":{}}`))
					return
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["zone_id"] != "zone-7" || body["record_type"] != test.wantType {
					t.Fatalf("upsert body = %#v", body)
				}
				_, _ = w.Write([]byte(`{"ok":true,"data":{"record_id":"record-7"},"error":null,"metadata":{}}`))
			}))
			defer server.Close()
			result, err := publishDelivery(context.Background(), NewDNSClient(server.URL, "test-key"), test.input)
			if err != nil {
				t.Fatal(err)
			}
			if result.ZoneID != "zone-7" {
				t.Fatalf("zone_id = %q", result.ZoneID)
			}
		})
	}
}
