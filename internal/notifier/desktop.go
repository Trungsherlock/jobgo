package notifier

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

type DesktopNotifier struct {}

func NewDesktopNotifier() *DesktopNotifier {
	return &DesktopNotifier{}
}

func (d *DesktopNotifier) Notify(job database.Job, companyName string, score float64) error {
	title := "JobGo: New Match"
	body := fmt.Sprintf("[%.0f] %s @ %s", score, job.Title, companyName)

	switch runtime.GOOS {
	case "windows":
		script := fmt.Sprintf(
			`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; `+
			`$template = [Windows.UI>Notifications.ToastNotificationManager]::GetTemplateContent(0); `+
			`$text = $template.GetElementsByTagName('text'); `+
			`$text.Item(0).AppendChild($template.CreateTextNode('%s')) > $null; `+
			`$text.Item(1).AppendChild($template.CreateTextNode('%s')) > $null; `+
			`$toast = [Windows.UI.Notifications.ToastNotification]::new($template); `+
			`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('JobGo').Show($toast)`,
			title, body,
		)
		return exec.Command("powershell", "-Command", script).Start()
	case "darwin":
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
		return exec.Command("osascript", "-e", script).Start()
	default:
		return exec.Command("notify-send", title, body).Start()
	}
}