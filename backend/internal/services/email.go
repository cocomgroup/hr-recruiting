package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// EmailService handles email sending
type EmailService struct {
	sendGridKey string
	fromEmail   string
	fromName    string
	client      *http.Client
}

// NewEmailService creates a new email service
func NewEmailService(sendGridKey string) *EmailService {
	return &EmailService{
		sendGridKey: sendGridKey,
		fromEmail:   "noreply@company.com",
		fromName:    "HR Recruiting",
		client:      &http.Client{},
	}
}

// SendApplicationConfirmation sends a confirmation email to the applicant
func (s *EmailService) SendApplicationConfirmation(email, firstName, jobID string) error {
	if s.sendGridKey == "" {
		log.Println("SendGrid API key not configured, skipping email")
		return nil
	}

	subject := "Application Received - Thank You for Applying!"
	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<h2>Thank you for your application, %s!</h2>
			<p>We've successfully received your application for the position.</p>
			<p>Our recruiting team will review your application and get back to you soon.</p>
			<p>In the meantime, you can:</p>
			<ul>
				<li>Track your application status in your dashboard</li>
				<li>Explore other open positions</li>
				<li>Connect with us on LinkedIn</li>
			</ul>
			<p>Best regards,<br>The Recruiting Team</p>
		</body>
		</html>
	`, firstName)

	return s.sendEmail(email, subject, htmlContent)
}

// SendStatusUpdate sends a status update email
func (s *EmailService) SendStatusUpdate(applicationID, status string) error {
	if s.sendGridKey == "" {
		log.Println("SendGrid API key not configured, skipping email")
		return nil
	}

	// In a real implementation, you would fetch the application details
	// from the database to get the candidate's email and job title
	log.Printf("Would send status update email for application %s: %s", applicationID, status)
	return nil
}

// SendInterviewInvitation sends an interview invitation
func (s *EmailService) SendInterviewInvitation(email, candidateName, jobTitle, interviewDate string) error {
	if s.sendGridKey == "" {
		log.Println("SendGrid API key not configured, skipping email")
		return nil
	}

	subject := fmt.Sprintf("Interview Invitation - %s", jobTitle)
	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<h2>Great news, %s!</h2>
			<p>We'd like to invite you for an interview for the <strong>%s</strong> position.</p>
			<p><strong>Interview Date:</strong> %s</p>
			<p>Please confirm your availability by replying to this email.</p>
			<p>We look forward to speaking with you!</p>
			<p>Best regards,<br>The Recruiting Team</p>
		</body>
		</html>
	`, candidateName, jobTitle, interviewDate)

	return s.sendEmail(email, subject, htmlContent)
}

// SendOfferLetter sends an offer letter
func (s *EmailService) SendOfferLetter(email, candidateName, jobTitle string) error {
	if s.sendGridKey == "" {
		log.Println("SendGrid API key not configured, skipping email")
		return nil
	}

	subject := fmt.Sprintf("Job Offer - %s", jobTitle)
	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<h2>Congratulations, %s!</h2>
			<p>We're excited to extend an offer for the <strong>%s</strong> position.</p>
			<p>Please review the attached offer letter and let us know if you have any questions.</p>
			<p>We look forward to welcoming you to our team!</p>
			<p>Best regards,<br>The Recruiting Team</p>
		</body>
		</html>
	`, candidateName, jobTitle)

	return s.sendEmail(email, subject, htmlContent)
}

// SendRejection sends a rejection email
func (s *EmailService) SendRejection(email, candidateName, jobTitle string) error {
	if s.sendGridKey == "" {
		log.Println("SendGrid API key not configured, skipping email")
		return nil
	}

	subject := fmt.Sprintf("Application Update - %s", jobTitle)
	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<p>Dear %s,</p>
			<p>Thank you for your interest in the <strong>%s</strong> position and for taking the time to apply.</p>
			<p>After careful consideration, we have decided to move forward with other candidates whose qualifications more closely match our current needs.</p>
			<p>We appreciate your interest in our company and encourage you to apply for future positions that match your skills and experience.</p>
			<p>We wish you the best in your job search.</p>
			<p>Best regards,<br>The Recruiting Team</p>
		</body>
		</html>
	`, candidateName, jobTitle)

	return s.sendEmail(email, subject, htmlContent)
}

// sendEmail sends an email using SendGrid API
func (s *EmailService) sendEmail(to, subject, htmlContent string) error {
	if s.sendGridKey == "" {
		return fmt.Errorf("SendGrid API key not configured")
	}

	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]string{
					{"email": to},
				},
			},
		},
		"from": map[string]string{
			"email": s.fromEmail,
			"name":  s.fromName,
		},
		"subject": subject,
		"content": []map[string]string{
			{
				"type":  "text/html",
				"value": htmlContent,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.sendGridKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("SendGrid returned status %d", resp.StatusCode)
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}