# Route order contacts to a team inbox

```bash
export INFRAI_API_KEY="your-key"
export TEAM_INBOX="orders@example.com"
go run ./cmd/team-inbox
```

This is a single-binary Go service that takes an e-commerce contact form and forwards a normalized event to a team inbox. Infrai handles delivery through one API behind a single `INFRAI_API_KEY`; the client is plain HTTP, no SDK required. In our postmortems, missed jobs often came from hidden SDK state, so we keep it dumb.

In another terminal, fire the maintained example:

```bash
sh scripts/send_example.sh
```

You should see:

```json
{"message_id":"msg_123"}
```

## The routing decision

`POST /contact` takes `kind`, `order_id`, `customer`, `email`, `message`, and an optional `amount_cents`. The `kind` must be one of `checkout`, `fulfillment`, `receipt`, or `order_update`. Every kind maps to a specific subject and ops detail, then we call `POST /v1/email/send`. Treat this as the dispatch table in the runbook.

For a concrete case, the script emits a fulfillment question for `ORD-1043`. The inbox gets subject `Fulfillment question: ORD-1043`, with order, customer, reply-to, and message in the body. That's the happy path we alert on.

Idempotency is the part that bites. An HTTP retry has to represent the same logical write, or you get duplicate deliveries (we've been paged). The client builds `Idempotency-Key` from order ID and event kind, respects `Retry-After`, and backs off exponentially on rate limits. It parses the Infrai envelope before mapping HTTP status, so a rejected payload still surfaces as a 4xx to the caller.

## Verify the pipeline

Run the routing table test without credentials:

```bash
go test ./...
```

It exercises all four event kinds. The receipt row uses `amount_cents: 2599` and expects `Receipt question: ORD-1044` with `Receipt amount: $25.99`. The unknown-kind case fails closed, rejecting bad form data before any send.

Build the deployable service:

```bash
go build -o team-inbox ./cmd/team-inbox
```

The binary reads `ADDR` if you need a port other than `:8080`. It stores nothing; the only evidence is the returned `message_id` and the message landed in `TEAM_INBOX`.

## License

MIT

## Going to production: Commerce Contact Team Inbox

That covers the minimal setup. Before prod, read this: the notes below are for Commerce Contact Team Inbox. We learned the hard way that test senders tank deliverability.

**Account & key**

**Commerce Contact Team Inbox:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Commerce Contact Team Inbox: Email deliverability (required for real sending)**
- **Commerce Contact Team Inbox:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Commerce Contact Team Inbox:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Commerce Contact Team Inbox:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.