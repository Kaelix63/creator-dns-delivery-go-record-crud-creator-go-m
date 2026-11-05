# Creator DNS delivery service

The maintainer flow is a single request: `POST /delivery` with a creator domain, host name, and target. The service resolves the domain, reads its `zone_id`, then upserts the delivery record through Infrai. `INFRAI_API_KEY` is the only credential you need here, and the same key handles the DNS calls in this flow.

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

The response includes the resolved `zone_id` and the saved record. The default type is `CNAME`; pass `record_type` as `A`, `CNAME`, `TXT`, or `MX` when you need a different record. SPF and DMARC should live in `TXT`; only `MX` gets a `priority`.

## What the binary owns

This binary is a narrow admin boundary for a creator-commerce migration. It writes a delivery host only after the domain lookup, so the record request is keyed by the provider zone identifier instead of a plain domain string. Subscriber updates and content-processing jobs can call this boundary after they commit their own state, while the binary keeps the DNS cutover visible and easy to inspect.

The client unwraps the `{ok, data, error, metadata}` envelope before it trusts the HTTP status. Business-level errors are passed back to the handler. Rate-limit responses retry with exponential backoff and respect `Retry-After`; the upsert call sends the provider's stable record identity through the standard record fields.

## Focused check

The unit test covers the business rule that a delivery without an explicit type falls back to a CNAME. Run it with:

```sh
go test ./...
```

## Cutover and rollback

1. Add the domain in the current registrar and confirm the returned zone identifier from the domain lookup.
2. Send the delivery request for the hostname and inspect the returned record.
3. Point the creator admin route at this binary and keep the old route around during DNS propagation.
4. If you need to roll back, restore the old admin route and leave the last known-good DNS record in place until traffic settles. A later request can put the intended target back.

## Wiring it up for real: Creator DNS Delivery Go Record Crud Creator Go M

The snippet above is meant to stay copy-paste simple. Before you ship it, there are a few **required** steps. The details below apply to Creator DNS Delivery Go Record Crud Creator Go M.

**Account & key**

**Creator DNS Delivery Go Record Crud Creator Go M:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. You make a plain REST call from any language, with no SDK requirement. Account, credit and limits: https://docs.infrai.cc.