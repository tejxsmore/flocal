package service

import (
	"context"
	"fmt"
	"html"

	"github.com/resend/resend-go/v2"

	"flocal/internal/config"
)

type EmailService interface {
	SendMagicLink(ctx context.Context, toEmail, toName, link string) error
	SendWelcome(ctx context.Context, toEmail, toName string) error
}

type resendEmailService struct {
	client    *resend.Client
	fromEmail string
	fromName  string
}

func NewResendEmailService(cfg config.ResendConfig) EmailService {
	return &resendEmailService{
		client:    resend.NewClient(cfg.APIKey),
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

func (s *resendEmailService) from() string {
	if s.fromName == "" {
		return s.fromEmail
	}

	return fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
}

func greetingFor(toName string) string {
	if toName == "" {
		return "Hi there"
	}

	return "Hi " + html.EscapeString(toName)
}

const emailShellTemplate = `
<body style="margin:0;padding:72px 20px;background-color:#232020;font-family:'Switzer',-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
	<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
		<tr>
			<td align="center">

				<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="max-width:420px;">

					<tr>
						<td align="center" style="padding-bottom:56px;">
							<span style="font-size:14px;line-height:1;font-weight:800;letter-spacing:-0.01em;color:#f4f4f4;">
								Flocal
							</span>
						</td>
					</tr>

					<tr>
						<td>
%s
						</td>
					</tr>

					%s

				</table>

			</td>
		</tr>
	</table>
</body>`

func emailShell(bodyHTML, footerHTML string) string {
	var footer string

	if footerHTML != "" {
		footer = fmt.Sprintf(`
					<tr>
						<td style="padding-top:64px;">
							%s
						</td>
					</tr>
				`, footerHTML)
	}

	return fmt.Sprintf(emailShellTemplate, bodyHTML, footer)
}

const baseFooterHTML = `
<span style="font-size:13px;line-height:1.7;color:#d8d8d8;">
	Questions? Just reply to this email — we're happy to help.
</span>
<br/><br/>
<span style="font-size:13px;line-height:1.7;color:#d8d8d8;">
	Regards,<br/>
	The Flocal Team
</span>`

func (s *resendEmailService) SendMagicLink(
	ctx context.Context,
	toEmail,
	toName,
	link string,
) error {
	body := fmt.Sprintf(`
		<h1 style="margin:0 0 20px;font-size:28px;line-height:1.25;font-weight:800;letter-spacing:-0.02em;color:#f4f4f4;">
			Sign in to Flocal
		</h1>

		<p style="margin:0 0 48px;font-size:15px;line-height:1.8;color:#eeeeee;">
			%s, click the button below to sign in. This link is valid for 15 minutes and can only be used once.
		</p>

		<table
			role="presentation"
			width="100%%"
			cellpadding="0"
			cellspacing="0"
			border="0"
			style="width:100%%;"
		>
			<tr>
				<td
					align="center"
					style="padding:0;border:1px solid #ff6803;border-radius:12px;background-color:#FF7315;"
				>
					<a
						href="%s"
						style="display:block;width:100%%;box-sizing:border-box;padding:18px 20px;font-size:15px;line-height:1.4;font-weight:800;color:#232020;text-decoration:none;text-align:center;"
					>
						Sign in to Flocal
					</a>
				</td>
			</tr>
		</table>

		<p style="margin:44px 0 0;font-size:13px;line-height:1.8;color:#d8d8d8;">
			Or copy and paste this link into your browser:
			<br/><br/>
			<span style="color:#f4f4f4;word-break:break-all;">%s</span>
		</p>

		<table
			role="presentation"
			width="100%%"
			cellpadding="0"
			cellspacing="0"
			border="0"
			style="width:100%%;margin-top:56px;"
		>
			<tr>
				<td
					style="border:1px solid #464040;border-radius:12px;background-color:#3A3535;padding:18px 20px;"
				>
					<span style="font-size:13px;line-height:1.6;font-weight:600;color:#eeeeee;">
						Didn't request this? You can safely ignore this email — your account stays secure.
					</span>
				</td>
			</tr>
		</table>
	`, greetingFor(toName), html.EscapeString(link), html.EscapeString(link))

	req := &resend.SendEmailRequest{
		From:    s.from(),
		To:      []string{toEmail},
		Subject: "Your Flocal sign-in link",
		Html:    emailShell(body, ""),
	}

	if _, err := s.client.Emails.SendWithContext(ctx, req); err != nil {
		return fmt.Errorf("service: send magic link email: %w", err)
	}

	return nil
}

func (s *resendEmailService) SendWelcome(
	ctx context.Context,
	toEmail,
	toName string,
) error {
	body := fmt.Sprintf(`
		<h1 style="margin:0 0 20px;font-size:28px;line-height:1.25;font-weight:800;letter-spacing:-0.02em;color:#f4f4f4;">
			Welcome to Flocal 🎉
		</h1>

		<p style="margin:0 0 48px;font-size:15px;line-height:1.8;color:#eeeeee;">
			%s, your account is ready. Here's how to get started:
		</p>

		<table
			role="presentation"
			width="100%%"
			cellpadding="0"
			cellspacing="0"
			border="0"
			style="width:100%%;border:1px solid #464040;border-radius:16px;background-color:#3A3535;"
		>
			<tr>
				<td style="padding:36px 28px;">
					<table
						role="presentation"
						width="100%%"
						cellpadding="0"
						cellspacing="0"
						border="0"
						style="width:100%%;"
					>
						<tr>
							<td
								width="38"
								valign="middle"
								style="width:38px;padding:0 0 34px 0;vertical-align:middle;"
							>
								<table
									role="presentation"
									width="36"
									height="36"
									cellpadding="0"
									cellspacing="0"
									border="0"
									style="width:36px;height:36px;"
								>
									<tr>
										<td
											align="center"
											valign="middle"
											width="32"
											height="32"
											style="width:32px;height:32px;padding:0;border:2px solid #ff6803;border-radius:50%%;background-color:#FF7315;font-size:14px;line-height:32px;font-weight:800;color:#232020;text-align:center;vertical-align:middle;"
										>
											1
										</td>
									</tr>
								</table>
							</td>

							<td
								valign="middle"
								style="padding:0 0 34px 16px;font-size:15px;line-height:1.6;color:#eeeeee;vertical-align:middle;"
							>
								Pick a topic from the spinner.
							</td>
						</tr>

						<tr>
							<td
								width="38"
								valign="middle"
								style="width:38px;padding:0 0 34px 0;vertical-align:middle;"
							>
								<table
									role="presentation"
									width="36"
									height="36"
									cellpadding="0"
									cellspacing="0"
									border="0"
									style="width:36px;height:36px;"
								>
									<tr>
										<td
											align="center"
											valign="middle"
											width="32"
											height="32"
											style="width:32px;height:32px;padding:0;border:1px solid #464040;border-radius:50%%;background-color:#1a1818;font-size:14px;line-height:32px;font-weight:800;color:#f4f4f4;text-align:center;vertical-align:middle;"
										>
											2
										</td>
									</tr>
								</table>
							</td>

							<td
								valign="middle"
								style="padding:0 0 34px 16px;font-size:15px;line-height:1.6;color:#eeeeee;vertical-align:middle;"
							>
								Speak for about 60 seconds.
							</td>
						</tr>

						<tr>
							<td
								width="38"
								valign="middle"
								style="width:38px;padding:0;vertical-align:middle;"
							>
								<table
									role="presentation"
									width="36"
									height="36"
									cellpadding="0"
									cellspacing="0"
									border="0"
									style="width:36px;height:36px;"
								>
									<tr>
										<td
											align="center"
											valign="middle"
											width="32"
											height="32"
											style="width:32px;height:32px;padding:0;border:1px solid #464040;border-radius:50%%;background-color:#1a1818;font-size:14px;line-height:32px;font-weight:800;color:#f4f4f4;text-align:center;vertical-align:middle;"
										>
											3
										</td>
									</tr>
								</table>
							</td>

							<td
								valign="middle"
								style="padding:0 0 0 16px;font-size:15px;line-height:1.6;color:#eeeeee;vertical-align:middle;"
							>
								Get instant AI feedback on how you did.
							</td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
	`, greetingFor(toName))

	req := &resend.SendEmailRequest{
		From:    s.from(),
		To:      []string{toEmail},
		Subject: "Welcome to Flocal",
		Html:    emailShell(body, baseFooterHTML),
	}

	if _, err := s.client.Emails.SendWithContext(ctx, req); err != nil {
		return fmt.Errorf("service: send welcome email: %w", err)
	}

	return nil
}
