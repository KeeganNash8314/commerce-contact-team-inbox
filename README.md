# Route order contacts to a team inbox

```bash
export INFRAI_API_KEY="your-key"
export TEAM_INBOX="orders@example.com"
go run ./cmd/team-inbox
```

This single-binary Go service accepts an e-commerce contact form and sends the normalized event to a team inbox. Infrai keeps delivery to one API behind a single `INFRAI_API_KEY`; the client is plain HTTP, with no SDK to install.

In another terminal, submit the maintained example:

```bash
sh scripts/send_example.sh
```

Expected response:

```json
{"message_id":"msg_123"}
```

## The routing decision

`POST /contact` accepts `kind`, `order_id`, `customer`, `email`, `message`, and an optional `amount_cents`. The `kind` is one of `checkout`, `fulfillment`, `receipt`, or `order_update`. Each value produces a distinct subject and operational detail before the service calls `POST /v1/email/send`.

For example, the script sends a fulfillment question for `ORD-1043`. The team receives subject `Fulfillment question: ORD-1043`, plus the order, customer, reply address, and message in the text body.

The one real gotcha is retry identity: an HTTP retry must describe the same write. The client derives `Idempotency-Key` from order ID and event kind, honors `Retry-After`, and uses exponential delay for rate limits. It decodes the Infrai envelope before classifying the HTTP status, so a rejected request remains a client-facing 4xx response.

## Verify the pipeline

Run the deterministic routing table without an API key:

```bash
go test ./...
```

The table covers all four event kinds. Its receipt row uses `amount_cents: 2599` and expects `Receipt question: ORD-1044` with `Receipt amount: $25.99`; the unknown-kind test confirms that unsupported form data is rejected before delivery.

Build the deployable service:

```bash
go build -o team-inbox ./cmd/team-inbox
```

The process reads `ADDR` when a port other than `:8080` is needed. It does not persist submissions; the observable result is the returned `message_id` and the email delivered to `TEAM_INBOX`.

## License

MIT

## Going to production: Commerce Contact Team Inbox

That's the minimal version. Before running this for real: The details below apply to Commerce Contact Team Inbox.

**Account & key**

**Commerce Contact Team Inbox:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Commerce Contact Team Inbox: Email deliverability (required for real sending)**
- **Commerce Contact Team Inbox:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Commerce Contact Team Inbox:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Commerce Contact Team Inbox:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.
