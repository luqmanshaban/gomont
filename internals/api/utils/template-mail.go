package utils

// Email templates for Gomont, matching the web client's visual identity:
// ink/paper/accent palette, hard borders, monospace for data values.
// Built with html/template (not text/template) so dynamic values like
// endpoint URLs and error strings are auto-escaped before being placed
// into HTML — these can contain characters that would otherwise break
// the markup or, in principle, be exploited if ever attacker-influenced.

const otpEmailHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
</head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Helvetica,Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4;padding:40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="480" cellpadding="0" cellspacing="0" style="max-width:480px;width:100%;">
          <tr>
            <td style="padding-bottom:28px;">
              <span style="font-size:22px;font-weight:900;color:#292929;letter-spacing:-0.5px;">Gomont</span>
              <span style="display:inline-block;width:8px;height:8px;background-color:#fe5722;margin-left:6px;"></span>
            </td>
          </tr>
          <tr>
            <td style="background-color:#ffffff;border:2px solid #292929;padding:36px 32px;">
              <table role="presentation" cellpadding="0" cellspacing="0" style="margin-bottom:14px;">
                <tr>
                  <td style="width:18px;height:1px;background-color:#fe5722;font-size:0;line-height:0;">&nbsp;</td>
                  <td style="padding-left:8px;font-family:'SF Mono',Consolas,monospace;font-size:12px;letter-spacing:1px;text-transform:uppercase;color:#5c5c5c;">Verify your email</td>
                </tr>
              </table>
              <h1 style="margin:0 0 12px;font-size:26px;font-weight:900;color:#292929;letter-spacing:-0.3px;">Your verification code</h1>
              <p style="margin:0 0 28px;font-size:15px;line-height:1.6;color:#5c5c5c;">
                Enter this code to continue. It expires in 15 minutes.
              </p>
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
                <tr>
                  <td align="center" style="background-color:#f4f4f4;border:2px solid #292929;padding:20px;">
                    <span style="font-family:'SF Mono',Consolas,monospace;font-size:32px;font-weight:700;letter-spacing:8px;color:#292929;">{{.Code}}</span>
                  </td>
                </tr>
              </table>
              <p style="margin:0;font-size:13px;line-height:1.6;color:#5c5c5c;">
                Didn't request this? You can safely ignore this email.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding-top:24px;text-align:center;">
              <p style="margin:0;font-size:12px;color:#5c5c5c;">Gomont &middot; Open-source uptime monitoring</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

const otpEmailText = `Your Gomont verification code is: {{.Code}}

Enter this code to continue. It expires in 15 minutes.

Didn't request this? You can safely ignore this email.

— Gomont`

const notificationEmailHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
</head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Helvetica,Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4;padding:40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="480" cellpadding="0" cellspacing="0" style="max-width:480px;width:100%;">
          <tr>
            <td style="padding-bottom:28px;">
              <span style="font-size:22px;font-weight:900;color:#292929;letter-spacing:-0.5px;">Gomont</span>
              <span style="display:inline-block;width:8px;height:8px;background-color:#fe5722;margin-left:6px;"></span>
            </td>
          </tr>
          <tr>
            <td style="background-color:#ffffff;border:2px solid #c4291c;padding:36px 32px;">
              <table role="presentation" cellpadding="0" cellspacing="0" style="margin-bottom:18px;">
                <tr>
                  <td style="border:1.5px solid #292929;padding:5px 10px;">
                    <span style="display:inline-block;width:8px;height:8px;background-color:#c4291c;margin-right:6px;vertical-align:middle;"></span>
                    <span style="font-family:'SF Mono',Consolas,monospace;font-size:12px;font-weight:700;letter-spacing:0.5px;text-transform:uppercase;color:#c4291c;vertical-align:middle;">Down</span>
                  </td>
                </tr>
              </table>
              <h1 style="margin:0 0 12px;font-size:24px;font-weight:900;color:#292929;letter-spacing:-0.3px;">A monitor went down</h1>
              <p style="margin:0 0 24px;font-size:15px;line-height:1.6;color:#5c5c5c;">
                One of your endpoints stopped responding as expected.
              </p>
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:28px;border:1.5px solid #d9d9d9;">
                <tr>
                  <td style="padding:14px 16px;border-bottom:1.5px solid #d9d9d9;font-family:'SF Mono',Consolas,monospace;font-size:11px;text-transform:uppercase;letter-spacing:0.5px;color:#5c5c5c;width:90px;">Endpoint</td>
                  <td style="padding:14px 16px;border-bottom:1.5px solid #d9d9d9;font-family:'SF Mono',Consolas,monospace;font-size:13px;color:#292929;word-break:break-all;">{{.URL}}</td>
                </tr>
                <tr>
                  <td style="padding:14px 16px;border-bottom:1.5px solid #d9d9d9;font-family:'SF Mono',Consolas,monospace;font-size:11px;text-transform:uppercase;letter-spacing:0.5px;color:#5c5c5c;">Error</td>
                  <td style="padding:14px 16px;border-bottom:1.5px solid #d9d9d9;font-family:'SF Mono',Consolas,monospace;font-size:13px;color:#c4291c;word-break:break-word;">{{.Err}}</td>
                </tr>
                <tr>
                  <td style="padding:14px 16px;font-family:'SF Mono',Consolas,monospace;font-size:11px;text-transform:uppercase;letter-spacing:0.5px;color:#5c5c5c;">Detected</td>
                  <td style="padding:14px 16px;font-family:'SF Mono',Consolas,monospace;font-size:13px;color:#292929;">{{.Time}}</td>
                </tr>
              </table>
              <table role="presentation" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="background-color:#292929;border:2px solid #292929;">
                    <a href="{{.DashboardURL}}" style="display:inline-block;padding:12px 24px;font-size:14px;font-weight:700;color:#f4f4f4;text-decoration:none;">View in dashboard &rarr;</a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding-top:24px;text-align:center;">
              <p style="margin:0;font-size:12px;color:#5c5c5c;">Gomont &middot; Open-source uptime monitoring</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

const notificationEmailText = `Gomont alert: a monitor went down

Endpoint: {{.URL}}
Error:    {{.Err}}
Detected: {{.Time}}

View in dashboard: {{.DashboardURL}}

— Gomont`