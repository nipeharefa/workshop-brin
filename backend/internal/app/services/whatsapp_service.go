package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"strings"

	"github.com/fajarAnd/workshop-brin/wa-service/internal/app/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsAppService interface {
	Start(ctx context.Context) error
	Stop() error
	SendMessage(ctx context.Context, phone, message string) error
	IsConnected() bool
	GetQRCode() (string, error)
	Logout() error
}

type whatsAppService struct {
	client            *whatsmeow.Client
	userService       UserService
	n8nService        N8NService
	flowiseService    FlowiseService
	workflowConfigSvc WorkflowConfigService
	dbPool            *pgxpool.Pool
	container         *sqlstore.Container
	device            *store.Device
	isConnected       bool
	qrCode            string
}

func NewWhatsAppService(userService UserService, n8nService N8NService, flowiseService FlowiseService, workflowConfigSvc WorkflowConfigService, dbPool *pgxpool.Pool) WhatsAppService {
	return &whatsAppService{
		userService:       userService,
		n8nService:        n8nService,
		flowiseService:    flowiseService,
		workflowConfigSvc: workflowConfigSvc,
		dbPool:            dbPool,
	}
}

func (s *whatsAppService) Start(ctx context.Context) error {
	log.Printf("[WhatsAppService] Starting WhatsApp service")

	// Convert pgx pool to database/sql compatible connection for whatsmeow
	dbConn := stdlib.OpenDBFromPool(s.dbPool)

	container := sqlstore.NewWithDB(dbConn, "postgres", nil)
	err := container.Upgrade(ctx)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to upgrade database schema: %v", err)
		return fmt.Errorf("failed to upgrade database schema: %w", err)
	}
	s.container = container

	// Get the first device (create if doesn't exist)
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to get device: %v", err)
		return fmt.Errorf("failed to get device: %w", err)
	}
	s.device = deviceStore

	// Create WhatsApp client
	s.client = whatsmeow.NewClient(deviceStore, nil)

	// Set up event handlers
	s.client.AddEventHandler(s.handleEvent)

	// Connect to WhatsApp
	if s.client.Store.ID == nil {
		log.Printf("[WhatsAppService] Device not registered, QR code will be generated on connect")
	}

	err = s.client.Connect()
	if err != nil {
		log.Printf("[WhatsAppService] Failed to connect to WhatsApp: %v", err)
		return fmt.Errorf("failed to connect to WhatsApp: %w", err)
	}

	log.Printf("[WhatsAppService] WhatsApp service started successfully")
	return nil
}

func (s *whatsAppService) Stop() error {
	log.Printf("[WhatsAppService] Stopping WhatsApp service")

	if s.client != nil {
		s.client.Disconnect()
		s.isConnected = false
	}

	log.Printf("[WhatsAppService] WhatsApp service stopped")
	return nil
}

func (s *whatsAppService) SendMessage(ctx context.Context, phone, message string) error {
	log.Printf("[WhatsAppService] Sending message to %s: %s", phone, message)

	if !s.isConnected {
		return fmt.Errorf("WhatsApp client not connected")
	}

	// Format phone number for WhatsApp JID
	jid, err := s.formatPhoneToJID(phone)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to format phone number %s: %v", phone, err)
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Create message
	msg := &waE2E.Message{
		Conversation: &message,
	}

	// Send message
	resp, err := s.client.SendMessage(ctx, jid, msg)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to send message to %s: %v", phone, err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("[WhatsAppService] Message sent successfully to %s (ID: %s)", phone, resp.ID)
	return nil
}

func (s *whatsAppService) IsConnected() bool {
	return s.isConnected
}

func (s *whatsAppService) GetQRCode() (string, error) {
	if s.qrCode == "" {
		return "", fmt.Errorf("QR code not available")
	}
	return s.qrCode, nil
}

func (s *whatsAppService) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		s.handleIncomingMessage(v)
	case *events.QR:
		s.handleQRCode(v)
	case *events.Connected:
		s.handleConnected(v)
	case *events.Disconnected:
		s.handleDisconnected(v)
	case *events.LoggedOut:
		s.handleLoggedOut(v)
	}
}

func (s *whatsAppService) handleIncomingMessage(evt *events.Message) {
	log.Printf("[WhatsAppService] Received message from %s", evt.Info.Sender.String())

	// Skip if message is from bot itself
	if evt.Info.IsFromMe {
		return
	}

	// Extract phone number from sender JID
	phone := s.extractPhoneFromJID(evt.Info.Sender)
	if phone == "" {
		log.Printf("[WhatsAppService] Failed to extract phone from JID: %s", evt.Info.Sender.String())
		return
	}

	ctx := context.Background()
	// Check user eligibility
	// TODO: Uncomment when you want spesific user AI Reply
	//eligible, err := s.userService.IsUserEligible(ctx, phone)
	//if err != nil {
	//	log.Printf("[WhatsAppService] Failed to check user eligibility for %s: %v", phone, err)
	//	return
	//}
	//
	//if !eligible {
	//	log.Printf("[WhatsAppService] User %s is not eligible, ignoring message", phone)
	//	// Optionally send a response to unregistered users
	//	//s.sendUnregisteredUserMessage(ctx, phone)
	//	return
	//}
	//
	//// Get user details
	//user, err := s.userService.GetUserByPhone(ctx, phone)
	//if err != nil {
	//	log.Printf("[WhatsAppService] Failed to get user by phone %s: %v", phone, err)
	//	return
	//}

	// Extract message text
	messageText := s.extractMessageText(evt.Message)
	if messageText == "" {
		log.Printf("[WhatsAppService] No text content in message from %s", phone)
		return
	}

	//log.Printf("[WhatsAppService] Processing message from %s (%s): %s", user.Name, phone, messageText)

	// Create user context
	userContext := &models.UserContext{
		UserID: uuid.New(),
		//Name:   user.Name,
		Name:  "Dummy",
		Phone: phone,
		Email: "dummy@email.com",
	}

	// Route message to appropriate workflow
	err := s.routeMessageToWorkflow(ctx, userContext, messageText)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to route message for user %s: %v", phone, err)
		// Send error message to user
		s.sendErrorMessage(ctx, phone)
		return
	}

	log.Printf("[WhatsAppService] Message routed to workflow successfully for user %s", phone)
}

func (s *whatsAppService) handleQRCode(evt *events.QR) {
	log.Printf("[WhatsAppService] QR code received, ready for scanning")
	s.qrCode = evt.Codes[0]
}

func (s *whatsAppService) handleConnected(_ *events.Connected) {
	log.Printf("[WhatsAppService] Connected to WhatsApp successfully")
	s.isConnected = true
	s.qrCode = ""
}

func (s *whatsAppService) handleDisconnected(_ *events.Disconnected) {
	log.Printf("[WhatsAppService] Disconnected from WhatsApp")
	s.isConnected = false
}

func (s *whatsAppService) handleLoggedOut(_ *events.LoggedOut) {
	log.Printf("[WhatsAppService] Logged out from WhatsApp")
	s.isConnected = false
}

func (s *whatsAppService) formatPhoneToJID(phone string) (types.JID, error) {
	// Remove non-numeric characters
	cleanPhone := strings.ReplaceAll(phone, "+", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, " ", "")

	if len(cleanPhone) == 0 {
		return types.JID{}, fmt.Errorf("invalid phone number")
	}

	// Create JID for individual chat
	jid := types.NewJID(cleanPhone, types.DefaultUserServer)
	return jid, nil
}

func (s *whatsAppService) extractPhoneFromJID(jid types.JID) string {
	return jid.User
}

func (s *whatsAppService) extractMessageText(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}

	// Extract text from different message types
	if msg.Conversation != nil {
		return *msg.Conversation
	}

	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		return *msg.ExtendedTextMessage.Text
	}

	// Add more message types as needed
	return ""
}

func (s *whatsAppService) sendUnregisteredUserMessage(ctx context.Context, phone string) {
	message := "Sorry, you are not registered to use this service. Please contact the administrator for access."
	err := s.SendMessage(ctx, phone, message)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to send unregistered user message to %s: %v", phone, err)
	}
}

func (s *whatsAppService) sendErrorMessage(ctx context.Context, phone string) {
	message := "Sorry, there was an error processing your message. Please try again later."
	err := s.SendMessage(ctx, phone, message)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to send error message to %s: %v", phone, err)
	}
}

func (s *whatsAppService) routeMessageToWorkflow(ctx context.Context, userContext *models.UserContext, message string) error {
	// Get global workflow configuration
	workflowType, err := s.workflowConfigSvc.GetActiveWorkflowType(ctx)
	if err != nil {
		log.Printf("[WhatsAppService] Failed to get workflow config: %v", err)
		workflowType = "n8n" // Default fallback
	}

	log.Printf("[WhatsAppService] Routing message to workflow: %s", workflowType)

	// Route to appropriate workflow
	switch workflowType {
	case "flowise":
		err = s.flowiseService.SendMessageToWorkflow(ctx, userContext, message)
		if err != nil {
			log.Printf("[WhatsAppService] Failed to send message to Flowise: %v", err)
			return fmt.Errorf("failed to send message to Flowise: %w", err)
		}
	case "n8n":
		err = s.n8nService.SendMessageToWorkflow(ctx, userContext, message)
		if err != nil {
			log.Printf("[WhatsAppService] Failed to send message to N8N: %v", err)
			return fmt.Errorf("failed to send message to N8N: %w", err)
		}
	default:
		log.Printf("[WhatsAppService] Unknown workflow type %s, defaulting to N8N", workflowType)
		err = s.n8nService.SendMessageToWorkflow(ctx, userContext, message)
		if err != nil {
			log.Printf("[WhatsAppService] Failed to send message to N8N (default): %v", err)
			return fmt.Errorf("failed to send message to N8N (default): %w", err)
		}
	}

	return nil
}

func (s *whatsAppService) Logout() error {
	log.Printf("[WhatsAppService] Logging out from WhatsApp")

	ctx := context.Background()

	if s.client != nil {
		if s.client.Store.ID != nil {
			err := s.client.Logout(ctx)
			if err != nil {
				log.Printf("[WhatsAppService] Warning: Failed to logout from WhatsApp server: %v", err)
			} else {
				log.Printf("[WhatsAppService] Successfully logged out from WhatsApp server")
			}
		} else {
			log.Printf("[WhatsAppService] Device not registered, skipping server logout")
		}

		s.client.Disconnect()
		s.isConnected = false
	}

	if s.device != nil {
		if s.device.ID != nil {
			err := s.device.Delete(ctx)
			if err != nil {
				log.Printf("[WhatsAppService] Warning: Failed to delete device from database: %v", err)
			} else {
				log.Printf("[WhatsAppService] Successfully deleted device from database")
			}
		} else {
			log.Printf("[WhatsAppService] Device JID not available, skipping database deletion")
		}
		s.device = nil
	}

	s.qrCode = ""
	log.Printf("[WhatsAppService] Logout process completed successfully")
	return nil
}
