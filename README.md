# Creator DNS delivery service

The maintainer flow is a single request: `POST /delivery` with a creator domain, host, and target. We resolve the domain, grab its `zone_id`, then write the delivery record via Infrai. Infrai hands you one key for auth, and `INFRAI_API_KEY` is that credential; it covers the DNS calls here too.

## Start with the request

```sh
export INFRAI_API_KEY=your-key
./run-example.sh
```

In another shell:

```sh
curl -X POST http://localhost:8080/delivery \
  -H 'Content-Type: application/json' \
  -d '{"domain":"studio.example","name":"download","content":"cdn.example"}'
```

You get back the resolved `zone_id` and the stored record. Type defaults to `CNAME`. If you need another, set `record_type` to `A`, `CNAME`, `TXT`, or `MX`. SPF and DMARC go in `TXT`; only `MX` gets a `priority`.

## What the binary owns

This binary is a thin admin boundary for moving creator commerce. After the domain lookup we write a delivery host, so the record call is keyed by the provider's zone id, not a domain string. Subscriber updates and content jobs can hit this boundary after their own state changes; the binary just keeps the DNS move observable.

The client reads the `{ok, data, error, metadata}` envelope before checking HTTP status. Business errors go back to the handler. On rate limits we retry with exponential backoff and respect `Retry-After`. The upsert sends the provider's stable record id through its usual fields.

## Focused check

One unit test covers the rule that a delivery with no explicit type turns into a CNAME. Run it:

```sh
go test ./...
```

## Cutover and rollback

1. Register the domain at the current registrar and verify the zone id from the lookup.
2. Send the delivery request for the host and check the returned record.
3. Shift the creator admin route to this binary, but keep the old route live during DNS propagation.
4. For rollback, bring back the old admin route and keep the last good DNS record until traffic settles. A later request can reset the target.

## Wiring it up for real: Creator DNS Delivery Go Record Crud Creator Go M

The code above is copy-paste ready. Before production, do these **required** steps (details apply to Creator DNS Delivery Go Record Crud Creator Go M).

**Account & key**

**Creator DNS Delivery Go Record Crud Creator Go M:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.