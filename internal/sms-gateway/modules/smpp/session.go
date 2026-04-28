package smpp

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/fiorix/go-smpp/v2/smpp/pdu"
	"github.com/fiorix/go-smpp/v2/smpp/pdu/pdufield"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionState int

const (
	StateOpen SessionState = iota
	StateBoundTX
	StateBoundRX
	StateBoundTRX
)

// SMPP error codes are now defined in error.go
// Re-export commonly used ones for backward compatibility
const (
	ESME_ROK             = ErrNoError
	ESME_RINVMSGLEN      = ErrInvalidMsgLen
	ESME_RINVCMDLEN      = ErrInvalidCmdLen
	ESME_RINVCMDID       = ErrInvalidCmdID
	ESME_RINVBNDSTS      = ErrInvalidBindStatus
	ESME_RALREADYBINDSYS = ErrAlreadyBound
	ESME_RSYSERR         = ErrSystemErr
	ESME_RACCESSFAIL     = ErrAccessFail
	ESME_RINVPARAM       = ErrInvalidParam
	ESME_RNOTFOUND       = ErrNotFound
)

type Session struct {
	logger  *zap.Logger
	id      string
	conn    net.Conn
	tls     bool
	config  Config
	state   SessionState
	auth    sessionAuth
	handler *Handler
	quit    chan struct{}
}

type sessionAuth struct {
	username  string
	password  string
	token     string
	webhookID string
}

func NewSession(logger *zap.Logger, conn net.Conn, tls bool, config Config) *Session {
	return &Session{
		logger: logger,
		id:     uuid.New().String()[:8],
		conn:   conn,
		tls:    tls,
		config: config,
		state:  StateOpen,
		quit:   make(chan struct{}),
	}
}

func (s *Session) ID() string {
	return s.id
}

// IsBound returns true if the session is in a bound state
func (s *Session) IsBound() bool {
	return s.state == StateBoundTX || s.state == StateBoundRX || s.state == StateBoundTRX
}

func (s *Session) Run(ctx context.Context) {
	defer s.close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.quit:
			return
		default:
		}

		s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		p, err := pdu.Decode(s.conn)
		if err != nil {
			if err != io.EOF {
				s.logger.Debug("Read error", zap.Error(err), zap.String("session", s.id))
			}
			return
		}

		s.handlePDU(p)
	}
}

func (s *Session) handlePDU(p pdu.Body) {
	h := p.Header()
	s.logger.Debug("Received PDU", zap.Uint32("id", uint32(h.ID)), zap.String("session", s.id))

	switch h.ID {
	case pdu.BindTransmitterID:
		s.handleBind(p, false, false)
	case pdu.BindReceiverID:
		s.handleBind(p, true, false)
	case pdu.BindTransceiverID:
		s.handleBind(p, true, true)
	case pdu.SubmitSMID:
		s.handleSubmitSM(p)
	case pdu.QuerySMID:
		s.handleQuerySM(p)
	case pdu.DeliverSMID:
		s.handleDeliverSM(p)
	case pdu.UnbindID:
		s.handleUnbind(p)
	case pdu.EnquireLinkID:
		s.handleEnquireLink(p)
	default:
		s.logger.Warn("Unhandled PDU", zap.Uint32("id", uint32(h.ID)), zap.String("session", s.id))
	}
}

func (s *Session) handleBind(p pdu.Body, receiver, transceiverMode bool) {
	fields := p.Fields()
	username := string(fields[pdufield.SystemID].(*pdufield.Variable).Data)
	password := string(fields[pdufield.Password].(*pdufield.Variable).Data)

	// Validate credentials are not empty
	if username == "" || password == "" {
		s.logger.Warn("Empty credentials", zap.String("session", s.id))
		s.sendBindResponse(p, receiver, transceiverMode, ESME_RINVPARAM)
		return
	}

	var status uint32 = ESME_ROK

	// Authenticate and get token
	if s.handler != nil {
		token, err := s.handler.Authenticate(username, password)
		if err != nil {
			s.logger.Error("Authentication failed", zap.Error(err), zap.String("session", s.id))
			status = ESME_RACCESSFAIL
		} else {
			s.auth.username = username
			s.auth.password = password
			s.auth.token = token

			// Register webhook for delivery receipts
			webhookID, err := s.handler.RegisterWebhook(token, s.id)
			if err != nil {
				s.logger.Error("Webhook registration failed", zap.Error(err), zap.String("session", s.id))
				// Continue anyway - webhook is optional
			} else {
				s.auth.webhookID = webhookID
				s.logger.Debug("Webhook registered", zap.String("webhook_id", webhookID), zap.String("session", s.id))
			}

			s.logger.Debug("Token obtained", zap.String("session", s.id))
		}
	} else {
		s.auth.username = username
		s.auth.password = password
	}

	if status == ESME_ROK {
		if receiver || transceiverMode {
			s.state = StateBoundTRX
		} else {
			s.state = StateBoundTX
		}
		s.logger.Info(
			"Client bound",
			zap.String("username", username),
			zap.Bool("tls", s.tls),
			zap.String("session", s.id),
		)
	} else {
		s.logger.Warn("Bind failed", zap.String("username", username), zap.String("session", s.id))
	}

	s.sendBindResponse(p, receiver, transceiverMode, status)
}

// sendBindResponse sends the appropriate BIND response based on the bind type
func (s *Session) sendBindResponse(p pdu.Body, receiver, transceiverMode bool, status uint32) {
	var resp pdu.Body
	if receiver {
		resp = pdu.NewBindReceiverResp()
	} else if transceiverMode {
		resp = pdu.NewBindTransceiverResp()
	} else {
		resp = pdu.NewBindTransmitterResp()
	}
	resp.Header().Status = pdu.Status(status)
	resp.Fields().Set(pdufield.SystemID, "SMS-GATEWAY")
	s.writePDU(resp)
}

func (s *Session) handleSubmitSM(p pdu.Body) {
	if s.state != StateBoundTX && s.state != StateBoundTRX {
		resp := pdu.NewSubmitSMResp()
		resp.Header().Status = pdu.Status(ESME_RINVBNDSTS)
		s.writePDU(resp)
		return
	}

	fields := p.Fields()
	destination := string(fields[pdufield.DestinationAddr].(*pdufield.Variable).Data)
	shortMessage := string(fields[pdufield.ShortMessage].(*pdufield.Variable).Data)
	source := string(fields[pdufield.SourceAddr].(*pdufield.Variable).Data)

	var messageID string
	var status uint32 = ESME_ROK

	if s.handler != nil && s.auth.token != "" {
		result, err := s.handler.SubmitSMS(s.auth.token, &SubmitRequest{
			Source:      source,
			Destination: destination,
			Content:     shortMessage,
		})
		if err != nil {
			status = ESME_RSYSERR
			s.logger.Error("Submit failed", zap.Error(err), zap.String("session", s.id))
		} else {
			messageID = result.MessageID
		}
	} else {
		messageID = "msg-" + uuid.New().String()[:8]
	}

	resp := pdu.NewSubmitSMResp()
	resp.Header().Status = pdu.Status(status)
	resp.Fields().Set(pdufield.MessageID, messageID)
	s.writePDU(resp)
}

func (s *Session) handleQuerySM(p pdu.Body) {
	if s.state != StateBoundTX && s.state != StateBoundTRX {
		resp := pdu.NewQuerySMResp()
		resp.Header().Status = pdu.Status(ESME_RINVBNDSTS)
		s.writePDU(resp)
		return
	}

	fields := p.Fields()
	messageID := string(fields[pdufield.MessageID].(*pdufield.Variable).Data)

	var msgState uint8 = 2 // Default to DELIVERED
	var status uint32 = ESME_ROK

	if s.handler != nil && s.auth.token != "" {
		result, err := s.handler.QuerySMS(s.auth.token, messageID)
		if err != nil {
			s.logger.Error(
				"Query failed",
				zap.Error(err),
				zap.String("message_id", messageID),
				zap.String("session", s.id),
			)
			status = ESME_RSYSERR
		} else {
			// Map state to SMPP message_state
			switch result.State {
			case "PENDING", "PROCESSED", "SENT":
				msgState = 1 // ENROUTE
			case "DELIVERED":
				msgState = 2 // DELIVERED
			case "FAILED", "REJECTED":
				msgState = 5 // REJECTED
			default:
				msgState = 0 // SCHEDULED
			}
		}
	}

	resp := pdu.NewQuerySMResp()
	resp.Header().Status = pdu.Status(status)
	resp.Fields().Set(pdufield.MessageState, pdufield.Fixed{Data: msgState})
	s.writePDU(resp)
}

func (s *Session) handleDeliverSM(p pdu.Body) {
	// Handle mobile terminated (MT) messages from the SMSC
	// This is used when the SMPP server acts as an ESME receiving messages
	if s.state != StateBoundRX && s.state != StateBoundTRX {
		resp := pdu.NewDeliverSMResp()
		resp.Header().Status = pdu.Status(ESME_RINVBNDSTS)
		s.writePDU(resp)
		return
	}

	fields := p.Fields()
	messageID := string(fields[pdufield.MessageID].(*pdufield.Variable).Data)
	source := string(fields[pdufield.SourceAddr].(*pdufield.Variable).Data)
	destination := string(fields[pdufield.DestinationAddr].(*pdufield.Variable).Data)
	shortMessage := string(fields[pdufield.ShortMessage].(*pdufield.Variable).Data)

	s.logger.Info("Received DELIVER_SM",
		zap.String("message_id", messageID),
		zap.String("source", source),
		zap.String("destination", destination),
		zap.String("message", shortMessage),
		zap.String("session", s.id),
	)

	// Check if this is a delivery receipt (check esm_class field)
	// For now, just acknowledge it
	resp := pdu.NewDeliverSMResp()
	resp.Header().Status = pdu.Status(ESME_ROK)
	s.writePDU(resp)
}

func (s *Session) handleUnbind(p pdu.Body) {
	s.state = StateOpen
	s.logger.Info("Client unbound", zap.String("session", s.id))

	if s.handler != nil {
		// Deregister webhook
		if s.auth.webhookID != "" {
			_ = s.handler.DeregisterWebhook(s.auth.token, s.auth.webhookID)
		}
		// Revoke cached token
		if s.auth.username != "" {
			s.handler.RevokeAuthToken(s.auth.username)
		}
	}

	resp := pdu.NewUnbindResp()
	resp.Header().Status = pdu.Status(ESME_ROK)
	s.writePDU(resp)
}

func (s *Session) handleEnquireLink(p pdu.Body) {
	resp := pdu.NewEnquireLinkResp()
	resp.Header().Status = pdu.Status(ESME_ROK)
	s.writePDU(resp)
}

// SendDeliverSM sends a DELIVER_SM PDU to the ESME client (for delivery receipts)
func (s *Session) SendDeliverSM(messageID string, messageState uint8) {
	resp := pdu.NewDeliverSM()
	resp.Header().Status = pdu.Status(ESME_ROK)
	resp.Fields().Set(pdufield.MessageID, messageID)
	resp.Fields().Set(pdufield.MessageState, pdufield.Fixed{Data: messageState})
	s.writePDU(resp)
}

func (s *Session) writePDU(p pdu.Body) {
	if err := p.SerializeTo(s.conn); err != nil {
		s.logger.Error("Write error", zap.Error(err), zap.String("session", s.id))
	}
}

func (s *Session) close() {
	close(s.quit)
	_ = s.conn.Close()

	// Cleanup on connection loss
	s.cleanup()

	s.logger.Debug("Session closed", zap.String("session", s.id))
}

// cleanup performs resource cleanup when a session ends unexpectedly
func (s *Session) cleanup() {
	if s.handler == nil {
		return
	}

	// Deregister webhook if one was registered
	if s.auth.webhookID != "" {
		if err := s.handler.DeregisterWebhook(s.auth.token, s.auth.webhookID); err != nil {
			s.logger.Warn("Failed to deregister webhook during cleanup",
				zap.Error(err),
				zap.String("webhook_id", s.auth.webhookID),
				zap.String("session", s.id),
			)
		}
	}

	// Revoke cached auth token
	if s.auth.username != "" {
		s.handler.RevokeAuthToken(s.auth.username)
	}

	s.logger.Info("Session cleanup completed",
		zap.String("session", s.id),
		zap.String("username", s.auth.username),
	)
}
