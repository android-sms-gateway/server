package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

// SMPP Command IDs
const (
	CmdBindTX          uint32 = 0x00000002
	CmdBindRX          uint32 = 0x00000004
	CmdBindTRX         uint32 = 0x00000006
	CmdBindTXResp      uint32 = 0x80000002
	CmdBindRXResp      uint32 = 0x80000004
	CmdBindTRXResp     uint32 = 0x80000006
	CmdSubmitSM        uint32 = 0x00000004
	CmdSubmitSMResp    uint32 = 0x80000004
	CmdDeliverSM       uint32 = 0x00000005
	CmdDeliverSMResp   uint32 = 0x80000005
	CmdQuerySM         uint32 = 0x00000003
	CmdQuerySMResp     uint32 = 0x80000003
	CmdUnbind          uint32 = 0x00000006
	CmdUnbindResp      uint32 = 0x80000006
	CmdEnquireLink     uint32 = 0x00000015
	CmdEnquireLinkResp uint32 = 0x80000015
)

// SMPP Status Codes
const (
	ESME_ROK uint32 = 0x00000000
)

var (
	address    = flag.String("addr", "localhost:2775", "SMPP server address")
	username   = flag.String("user", "", "Username for authentication")
	password   = flag.String("pass", "", "Password for authentication")
	systemType = flag.String("systype", "", "System type for bind")
	timeout    = flag.Duration("timeout", 10*time.Second, "Connection timeout")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "SMPP Test Client - Simple tool to test SMPP server implementation\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command> [command-args...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  bind-tx       Bind as transmitter\n")
		fmt.Fprintf(os.Stderr, "  bind-rx       Bind as receiver\n")
		fmt.Fprintf(os.Stderr, "  bind-trx      Bind as transceiver\n")
		fmt.Fprintf(os.Stderr, "  send-sms      Send SMS (binds, sends, unbinds)\n")
		fmt.Fprintf(os.Stderr, "  query         Query message status\n")
		fmt.Fprintf(os.Stderr, "  enquire       Send enquire link\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -user testuser -pass testpass bind-tx\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -user testuser -pass testpass send-sms 1234567890 \"Hello World\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -user testuser -pass testpass query msg-12345678\n", os.Args[0])
	}

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "bind-tx":
		runBindTransmitter()
	case "bind-rx":
		runBindReceiver()
	case "bind-trx":
		runBindTransceiver()
	case "send-sms":
		runSendSMS(cmdArgs)
	case "query":
		runQuery(cmdArgs)
	case "enquire":
		runEnquireLink()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

// connect establishes a TCP connection to the SMPP server
func connect() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", *address, *timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", *address, err)
	}
	fmt.Printf("✓ Connected to %s\n", *address)
	return conn, nil
}

// writePDU sends a raw SMPP PDU to the server
func writePDU(conn net.Conn, cmdID, seq uint32, body []byte) error {
	// PDU format: length(4) + command_id(4) + status(4) + seq(4) + body
	length := uint32(12 + len(body))

	buf := make([]byte, 0, length)

	// Length
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, length)
	buf = append(buf, lenBytes...)

	// Command ID
	cmdBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cmdBytes, cmdID)
	buf = append(buf, cmdBytes...)

	// Status (0 = success for requests)
	statusBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(statusBytes, 0)
	buf = append(buf, statusBytes...)

	// Sequence number
	seqBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(seqBytes, seq)
	buf = append(buf, seqBytes...)

	// Body
	buf = append(buf, body...)

	_, err := conn.Write(buf)
	return err
}

// readPDU reads a raw SMPP PDU from the connection
func readPDU(conn net.Conn) (cmdID uint32, status uint32, seq uint32, body []byte, err error) {
	// Read length
	lenBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, lenBuf)
	if err != nil {
		return
	}
	length := binary.BigEndian.Uint32(lenBuf)

	if length < 12 || length > 8192 {
		err = fmt.Errorf("invalid PDU length: %d", length)
		return
	}

	// Read rest of PDU
	buf := make([]byte, length-4)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return
	}

	cmdID = binary.BigEndian.Uint32(buf[0:4])
	status = binary.BigEndian.Uint32(buf[4:8])
	seq = binary.BigEndian.Uint32(buf[8:12])
	body = buf[12:]
	return
}

// createBindBody creates the body for a bind PDU
func createBindBody() []byte {
	// Bind body: system_id(C-string) + password(C-string) + system_type(C-string) +
	// interface_version(int8) + addr_ton(int8) + addr_npi(int8) + address_range(C-string)

	var buf []byte
	buf = append(buf, []byte(*username)...)
	buf = append(buf, 0) // null terminator

	buf = append(buf, []byte(*password)...)
	buf = append(buf, 0)

	buf = append(buf, []byte(*systemType)...)
	buf = append(buf, 0)

	buf = append(buf, 0x34) // interface version (SMPP 3.4)
	buf = append(buf, 1)    // addr_ton (International)
	buf = append(buf, 1)    // addr_npi (ISDN)

	buf = append(buf, 0) // address_range (empty)

	return buf
}

func runBindTransmitter() {
	fmt.Printf("Connecting to SMPP server at %s as transmitter...\n", *address)

	conn, err := connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	seq := uint32(time.Now().UnixNano())
	if err := writePDU(conn, CmdBindTX, seq, createBindBody()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send bind: %v\n", err)
		os.Exit(1)
	}

	cmdID, status, _, body, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(1)
	}

	if cmdID != CmdBindTXResp {
		fmt.Fprintf(os.Stderr, "Unexpected response: 0x%08X\n", cmdID)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Bind failed with status: 0x%08X\n", status)
		os.Exit(1)
	}

	// Parse system_id from response
	if len(body) > 0 {
		systemID := string(body[:findNull(body)])
		fmt.Printf("✓ Bound as transmitter successfully (System: %s)\n", systemID)
	} else {
		fmt.Println("✓ Bound as transmitter successfully")
	}

	fmt.Println("Press Ctrl+C to disconnect")
	select {}
}

func runBindReceiver() {
	fmt.Printf("Connecting to SMPP server at %s as receiver...\n", *address)

	conn, err := connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	seq := uint32(time.Now().UnixNano())
	if err := writePDU(conn, CmdBindRX, seq, createBindBody()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send bind: %v\n", err)
		os.Exit(1)
	}

	cmdID, status, _, body, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(1)
	}

	if cmdID != CmdBindRXResp {
		fmt.Fprintf(os.Stderr, "Unexpected response: 0x%08X\n", cmdID)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Bind failed with status: 0x%08X\n", status)
		os.Exit(1)
	}

	if len(body) > 0 {
		systemID := string(body[:findNull(body)])
		fmt.Printf("✓ Bound as receiver successfully (System: %s)\n", systemID)
	} else {
		fmt.Println("✓ Bound as receiver successfully")
	}

	fmt.Println("Press Ctrl+C to disconnect")
	select {}
}

func runBindTransceiver() {
	fmt.Printf("Connecting to SMPP server at %s as transceiver...\n", *address)

	conn, err := connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	seq := uint32(time.Now().UnixNano())
	if err := writePDU(conn, CmdBindTRX, seq, createBindBody()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send bind: %v\n", err)
		os.Exit(1)
	}

	cmdID, status, _, body, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(1)
	}

	if cmdID != CmdBindTRXResp {
		fmt.Fprintf(os.Stderr, "Unexpected response: 0x%08X\n", cmdID)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Bind failed with status: 0x%08X\n", status)
		os.Exit(1)
	}

	if len(body) > 0 {
		systemID := string(body[:findNull(body)])
		fmt.Printf("✓ Bound as transceiver successfully (System: %s)\n", systemID)
	} else {
		fmt.Println("✓ Bound as transceiver successfully")
	}

	fmt.Println("Press Ctrl+C to disconnect")
	select {}
}

func runSendSMS(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: send-sms <destination> <message>\n")
		os.Exit(1)
	}

	destination := args[0]
	message := args[1]

	fmt.Printf("Connecting to SMPP server at %s...\n", *address)

	conn, err := connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Bind as transmitter
	seq := uint32(time.Now().UnixNano())
	if err := writePDU(conn, CmdBindTX, seq, createBindBody()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send bind: %v\n", err)
		os.Exit(1)
	}

	_, status, _, _, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read bind response: %v\n", err)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Bind failed: 0x%08X\n", status)
		os.Exit(1)
	}
	fmt.Println("✓ Bound as transmitter successfully")

	// Create submit_sm body
	submitBody := createSubmitSMBody("", destination, message)

	seq = uint32(time.Now().UnixNano())
	fmt.Printf("Sending SMS to %s: %s\n", destination, message)

	if err := writePDU(conn, CmdSubmitSM, seq, submitBody); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send submit_sm: %v\n", err)
		os.Exit(1)
	}

	cmdID, status, _, body, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read submit_sm response: %v\n", err)
		os.Exit(1)
	}

	if cmdID != CmdSubmitSMResp {
		fmt.Fprintf(os.Stderr, "Unexpected response: 0x%08X\n", cmdID)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Submit failed: 0x%08X\n", status)
		os.Exit(1)
	}

	msgID := string(body[:findNull(body)])
	fmt.Printf("✓ Message submitted successfully! Message ID: %s\n", msgID)

	// Unbind
	fmt.Println("Unbinding...")
	seq = uint32(time.Now().UnixNano())
	writePDU(conn, CmdUnbind, seq, nil)

	cmdID, status, _, _, _ = readPDU(conn)
	if cmdID == CmdUnbindResp && status == ESME_ROK {
		fmt.Println("✓ Disconnected")
	}
}

func createSubmitSMBody(serviceType, destination, message string) []byte {
	// submit_sm body format:
	// service_type(C-string) + source_addr_ton(int8) + source_addr_npi(int8) +
	// source_addr(C-string) + dest_addr_ton(int8) + dest_addr_npi(int8) +
	// destination_addr(C-string) + esm_class(int8) + protocol_id(int8) +
	// priority_flag(int8) + schedule_delivery_time(C-string) + validity_period(C-string) +
	// registered_delivery(int8) + replace_if_cpresent_flag(int8) + data_coding(int8) +
	// sm_default_msg_id(int8) + sm_length(int8) + sm(C-string)

	var buf []byte

	// service_type
	buf = append(buf, []byte(serviceType)...)
	buf = append(buf, 0)

	// source_addr_ton, source_addr_npi, source_addr
	buf = append(buf, 1) // source_addr_ton (International)
	buf = append(buf, 1) // source_addr_npi (ISDN)
	buf = append(buf, 0) // source_addr (empty)

	// dest_addr_ton, dest_addr_npi, destination_addr
	buf = append(buf, 1) // dest_addr_ton (International)
	buf = append(buf, 1) // dest_addr_npi (ISDN)
	buf = append(buf, []byte(destination)...)
	buf = append(buf, 0)

	// esm_class, protocol_id, priority_flag
	buf = append(buf, 0) // esm_class
	buf = append(buf, 0) // protocol_id
	buf = append(buf, 0) // priority_flag

	// schedule_delivery_time, validity_period
	buf = append(buf, 0) // schedule_delivery_time (empty)
	buf = append(buf, 0) // validity_period (empty)

	// registered_delivery, replace_if_cpresent_flag, data_coding, sm_default_msg_id
	buf = append(buf, 0) // registered_delivery
	buf = append(buf, 0) // replace_if_cpresent_flag
	buf = append(buf, 0) // data_coding
	buf = append(buf, 0) // sm_default_msg_id

	// sm_length, sm
	buf = append(buf, byte(len(message)))
	buf = append(buf, []byte(message)...)

	return buf
}

func runQuery(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: query <message_id>\n")
		os.Exit(1)
	}

	messageID := args[0]

	fmt.Printf("Connecting to SMPP server at %s...\n", *address)

	conn, err := connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Bind as transceiver
	seq := uint32(time.Now().UnixNano())
	if err := writePDU(conn, CmdBindTRX, seq, createBindBody()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send bind: %v\n", err)
		os.Exit(1)
	}

	_, status, _, _, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read bind response: %v\n", err)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Bind failed: 0x%08X\n", status)
		os.Exit(1)
	}
	fmt.Println("✓ Bound as transceiver successfully")

	// Create query_sm body: message_id(C-string)
	var queryBody []byte
	queryBody = append(queryBody, []byte(messageID)...)
	queryBody = append(queryBody, 0)

	seq = uint32(time.Now().UnixNano())
	fmt.Printf("Querying message status for ID: %s\n", messageID)

	if err := writePDU(conn, CmdQuerySM, seq, queryBody); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send query_sm: %v\n", err)
		os.Exit(1)
	}

	cmdID, status, _, body, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read query_sm response: %v\n", err)
		os.Exit(1)
	}

	if cmdID != CmdQuerySMResp {
		fmt.Fprintf(os.Stderr, "Unexpected response: 0x%08X\n", cmdID)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Query failed: 0x%08X\n", status)
		os.Exit(1)
	}

	// Parse response: final_date(C-string) + message_state(int8) + error_code(int8)
	offset := findNull(body)
	finalDate := string(body[:offset])
	if len(body) > offset+1 {
		offset++
		messageState := body[offset]
		var errorCode byte
		if len(body) > offset+1 {
			errorCode = body[offset+1]
		}
		fmt.Printf("✓ Query successful!\n")
		fmt.Printf("  Final Date: %s\n", finalDate)
		fmt.Printf("  Message State: %d\n", messageState)
		fmt.Printf("  Error Code: %d\n", errorCode)
	} else {
		fmt.Printf("✓ Query successful! Final Date: %s\n", finalDate)
	}

	// Unbind
	fmt.Println("Unbinding...")
	seq = uint32(time.Now().UnixNano())
	writePDU(conn, CmdUnbind, seq, nil)

	cmdID, status, _, _, _ = readPDU(conn)
	if cmdID == CmdUnbindResp && status == ESME_ROK {
		fmt.Println("✓ Disconnected")
	}
}

func runEnquireLink() {
	fmt.Printf("Connecting to SMPP server at %s...\n", *address)

	conn, err := connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Bind first
	seq := uint32(time.Now().UnixNano())
	if err := writePDU(conn, CmdBindTRX, seq, createBindBody()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send bind: %v\n", err)
		os.Exit(1)
	}

	_, status, _, _, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read bind response: %v\n", err)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Bind failed: 0x%08X\n", status)
		os.Exit(1)
	}
	fmt.Println("✓ Bound as transceiver successfully")

	// Send enquire link
	fmt.Println("Sending enquire link...")
	seq = uint32(time.Now().UnixNano())
	if err := writePDU(conn, CmdEnquireLink, seq, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send enquire_link: %v\n", err)
		os.Exit(1)
	}

	cmdID, status, _, _, err := readPDU(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read enquire_link response: %v\n", err)
		os.Exit(1)
	}

	if cmdID != CmdEnquireLinkResp {
		fmt.Fprintf(os.Stderr, "Unexpected response: 0x%08X\n", cmdID)
		os.Exit(1)
	}

	if status != ESME_ROK {
		fmt.Fprintf(os.Stderr, "Enquire link failed: 0x%08X\n", status)
		os.Exit(1)
	}

	fmt.Println("✓ Enquire link successful - server is alive")

	// Unbind
	fmt.Println("Unbinding...")
	seq = uint32(time.Now().UnixNano())
	writePDU(conn, CmdUnbind, seq, nil)

	cmdID, status, _, _, _ = readPDU(conn)
	if cmdID == CmdUnbindResp && status == ESME_ROK {
		fmt.Println("✓ Disconnected")
	}
}

// findNull finds the position of the first null byte in a byte slice
func findNull(data []byte) int {
	for i, b := range data {
		if b == 0 {
			return i
		}
	}
	return len(data)
}
