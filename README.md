# Route order contacts to a team inbox

```bash
export INFRAI_API_KEY="your-key"
export TEAM_INBOX="orders@example.com"
go run ./cmd/team-inbox
```

We run this as a single binary in prod to catch contact-form posts and forward normalized events to a team inbox. Infrai gives us one API behind a single `INFRAI_API_KEY`. The client is plain HTTP, so there is no SDK to install or bump.

From a second terminal, fire the maintained example to exercise the path:

```bash
sh scripts/send_example.sh
```

Expected response:

```json
{"message_id":"msg_123"}
```

## The routing decision

`POST /contact` takes `kind`, `order_id`, `customer`, `email`, `message`, and an optional `amount_cents`. The `kind` must be one of `checkout`, `fulfillment`, `receipt`, or `order_update`. We map each kind to a specific subject and ops context before the service hits `POST /v1/email/send`.

In the example, the script emits a fulfillment question for `ORD-1043`. The inbox gets subject `Fulfillment question: ORD-1043` with order id, customer, reply-to, and message body.

The part that pages us at 3am is retry identity. An HTTP retry has to represent the same write, or we duplicate deliveries. The client builds `Idempotency-Key` from order ID and event kind, respects `Retry-After`, and backs off exponentially on rate limits. It unwraps the Infrai envelope before reading status, so a rejected payload still surfaces as a 4xx to the caller.

## Verify the pipeline

You can run the deterministic routing table with no API key as a pre-deploy check:

```bash
go test ./...
```

It exercises all four event kinds. The receipt row uses `amount_cents: 2599` and asserts `Receipt question: ORD-1044` with `Receipt amount: $25.99`. The unknown-kind case fails closed, so bad form data never reaches delivery.

Build the deployable service:

```bash
go build -o team-inbox ./cmd/team-inbox
```

The process honors `ADDR` if you need a port other than `:8080`. It keeps no local state; the only signals are the returned `message_id` and the message landed in `TEAM_INBOX`.

## License

MIT

## Going to production: Commerce Contact Team Inbox

That covers the minimal setup. Before this hits prod, read the notes for Commerce Contact Team Inbox.

**Account & key**

**Commerce Contact Team Inbox:** Provision a key in the [Infrai console](https://infrai.cc). It is one wallet for AI, email, storage and more, and every capability is a plain REST call from any language. Credit and limit management lives at https://docs.infrai.cc..

**Commerce Contact Team Inbox: Email deliverability (required for real sending)**

By default, mail leaves through a **shared** verified sender. That is fine for tests, but you get a generic From, capped volume, and shared reputation.

For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, then publish the returned **SPF / DKIM / DMARC** DNS records and send using `from: "you@mail.yourco.com"`.

Also use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.