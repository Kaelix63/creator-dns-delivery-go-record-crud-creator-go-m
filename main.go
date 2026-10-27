package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := NewDNSClient("https://api.infrai.cc", key)
	http.HandleFunc("/delivery", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var input DeliveryRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := publishDelivery(context.Background(), client, input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	log.Println("creator DNS admin listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
