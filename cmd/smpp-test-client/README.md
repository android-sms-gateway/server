# SMPP Test Client

A simple command-line tool to test the SMPP server implementation.

## Building

```bash
# Build the test client
go build -o smpp-test-client ./cmd/smpp-test-client/
```

## Usage

```bash
./smpp-test-client [options] <command> [command-args...]
```

### Options

| Option     | Description                 | Default                      |
| ---------- | --------------------------- | ---------------------------- |
| `-addr`    | SMPP server address         | `localhost:2775`             |
| `-user`    | Username for authentication | (required for most commands) |
| `-pass`    | Password for authentication | (required for most commands) |
| `-systype` | System type for bind        | (empty)                      |
| `-timeout` | Connection timeout          | `10s`                        |

### Commands

#### bind-tx
Bind to the server as a transmitter. This mode allows sending SMS messages.

```bash
./smpp-test-client -user testuser -pass testpass bind-tx
```

#### bind-rx
Bind to the server as a receiver. This mode allows receiving SMS messages and delivery receipts.

```bash
./smpp-test-client -user testuser -pass testpass bind-rx
```

#### bind-trx
Bind to the server as a transceiver. This mode combines both transmitter and receiver capabilities.

```bash
./smpp-test-client -user testuser -pass testpass bind-trx
```

#### send-sms
Send an SMS message. This command automatically binds as transmitter, sends the message, and unbinds.

```bash
./smpp-test-client -user testuser -pass testpass send-sms 1234567890 "Hello World"
```

Arguments:
- `destination` - Destination phone number
- `message` - SMS message text

#### query
Query the status of a previously sent message.

```bash
./smpp-test-client -user testuser -pass testpass query msg-12345678
```

Arguments:
- `message_id` - Message ID returned from send-sms

#### enquire
Send an enquire link to verify the server is alive and responsive.

```bash
./smpp-test-client -user testuser -pass testpass enquire
```

## Examples

### Full workflow: Send and query an SMS

```bash
# 1. Send an SMS
./smpp-test-client -user testuser -pass testpass send-sms 1234567890 "Hello World"
# Output: ✓ Message submitted successfully! Message ID: msg-abc123

# 2. Query the message status
./smpp-test-client -user testuser -pass testpass query msg-abc123
# Output: ✓ Query successful!
#           Final Date: 
#           Message State: 2
#           Error Code: 0
```

### Testing with TLS

```bash
# Connect to TLS-enabled SMPP server (port 2776)
./smpp-test-client -addr localhost:2776 -user testuser -pass testpass bind-tx
```

### Testing connection timeout

```bash
# Test with custom timeout
./smpp-test-client -addr localhost:2775 -user testuser -pass testpass -timeout 5s bind-tx
```

## Protocol Details

The test client implements SMPP v3.4 protocol with the following settings:
- Interface Version: 0x34 (v3.4)
- Address TON: 1 (International)
- Address NPI: 1 (ISDN/telephone numbering plan)

### Supported SMPP Operations

| Operation    | Command ID | Description          |
| ------------ | ---------- | -------------------- |
| bind-tx      | 0x00000002 | Bind as transmitter  |
| bind-rx      | 0x00000004 | Bind as receiver     |
| bind-trx     | 0x00000006 | Bind as transceiver  |
| submit_sm    | 0x00000004 | Submit short message |
| query_sm     | 0x00000003 | Query message status |
| unbind       | 0x00000006 | Unbind from server   |
| enquire_link | 0x00000015 | Enquire link         |

## Troubleshooting

### Connection refused
Make sure the SMPP server is running and listening on the specified address.

```bash
# Check if server is running
netstat -an | grep 2775
```

### Authentication failed
Verify the username and password are correct. The server validates credentials against the configured authentication backend.

### Bind failed with status 0x0000000F (ESME_RALREADYBINDSYS)
The session is already bound. This can happen if you try to bind twice on the same connection.

### Timeout
Increase the timeout value or check network connectivity to the server.

```bash
./smpp-test-client -timeout 30s -user testuser -pass testpass bind-tx
```
