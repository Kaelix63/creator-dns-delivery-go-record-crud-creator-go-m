package main

import (
	"context"
	"fmt"
	"strings"
)

type Zone struct {
	ZoneID string `json:"zone_id"`
}

type DeliveryRequest struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Type    string `json:"record_type"`
}

type DeliveryResult struct {
	ZoneID string         `json:"zone_id"`
	Record map[string]any `json:"record"`
}

func publishDelivery(ctx context.Context, client *DNSClient, input DeliveryRequest) (DeliveryResult, error) {
	if input.Domain == "" || input.Name == "" || input.Content == "" {
		return DeliveryResult{}, fmt.Errorf("domain, name, and content are required")
	}
	if input.Type == "" {
		input.Type = "CNAME"
	}
	if input.Type != "A" && input.Type != "CNAME" && input.Type != "TXT" && input.Type != "MX" {
		return DeliveryResult{}, fmt.Errorf("unsupported record_type")
	}
	var zone Zone
	if err := client.DomainGet(ctx, input.Domain, &zone); err != nil {
		return DeliveryResult{}, err
	}
	if strings.TrimSpace(zone.ZoneID) == "" {
		return DeliveryResult{}, fmt.Errorf("domain response did not include zone_id")
	}
	record := map[string]any{"zone_id": zone.ZoneID, "record_type": input.Type, "name": input.Name, "content": input.Content}
	if input.Type == "MX" {
		record["priority"] = 10
	}
	var saved map[string]any
	if err := client.RecordUpsert(ctx, record, &saved); err != nil {
		return DeliveryResult{}, err
	}
	return DeliveryResult{ZoneID: zone.ZoneID, Record: saved}, nil
}
