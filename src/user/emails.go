package user

import (
	"fmt"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

func SendInviteEmail(nickname string, email string, link string) error {
	return utils.SendEmail(
		[]string{email},
		"Cosmos Invitation for "+nickname,
		fmt.Sprintf(`<h1>You have been invited!</h1>
Hello %s, <br>
The admin of a Cosmos Server invited you to join their server. <br>
In order to join, you can click the following link to setup your account: <br>
<a class="button" href="%s">Setup</a> <br><br>
See you soon!! <br>
`, nickname, link))
}

func SendAdminPasswordEmail(nickname string, email string, link string) error {
	return utils.SendEmail(
		[]string{email},
		"Cosmos Password Reset",
		fmt.Sprintf(`<h1>Password Reset</h1>
Hello %s, <br>
The admin of a Cosmos Server has sent you a password reset link. <br>
In order to reset your password, you can click the following link and fill in the form: <br>
<a class="button" href="%s">Reset Password</a> <br><br>
See you soon!! <br>
`, nickname, link))
}

func SendPasswordEmail(nickname string, email string, link string) error {
	return utils.SendEmail(
		[]string{email},
		"Cosmos Password Reset",
		fmt.Sprintf(`<h1>Password Reset</h1>
Hello %s, <br>
You have requested a password reset. If it wasn't you, please alert your server admin. <br>
If it was you, you can click the following link and fill in the form: <br>
<a class="button" href="%s">Reset Password</a> <br><br>
See you soon!! <br>
`, nickname, link))
}

func SendLoginNotificationEmail(nickname string, email string, ip string, date time.Time) error {
	return utils.SendEmail(
		[]string{email},
		"Cosmos Login Notification",
		fmt.Sprintf(`<h1>Login Notification</h1>
Hello %s, <br>
Your account has been logged into. If it wasn't you, please reset your password and alert your server admin. <br>
If it was you, you can ignore this email. <br><br>
The login was from the following IP: %s <br>
On the following date: %s <br><br>
`, nickname, ip, date.Format("2006-01-02 15:04:05")))
}
