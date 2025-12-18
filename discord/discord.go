package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/kolesaev/alertmanager-discord/alertmanager"
	"github.com/kolesaev/alertmanager-discord/config"
)

// SendAlerts deals with the macro logic of sending alerts to Discord Channels
func SendAlerts(
	discordChannelName string,
	alertmanagerBody alertmanager.MessageBody,
	configs config.Config) error {

	alertmanagerBodyInfo := alertmanager.ExtractBodyInfo(alertmanagerBody, configs)

	discordChannel, err := getDiscordChannel(discordChannelName, configs)
	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error trying to get Discord Channel \n%+v", err)
	}

	if alertmanager.CheckIfHasOnlySeveritiesToIgnoreWhenAlone(
		alertmanagerBodyInfo.CountBySeverity,
		discordChannel, configs) {

		return fmt.Errorf(
			`discord.SendAlerts: There are only alerts with severities to be ignored, message not sent.
			Severity Count: %+v`,
			alertmanagerBodyInfo.CountBySeverity)

	}

	discordMessage, err := createDiscordMessage(alertmanagerBodyInfo, discordChannel, configs)
	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error trying to create Discord Message \n%+v", err)
	}

	jsonDiscordMessage, err := json.Marshal(discordMessage)
	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error Marshaling Discord Message \n%+v", err)
	}

	r, err := http.Post(
		discordChannel.WebhookURL,
		"application/json",
		bytes.NewReader(jsonDiscordMessage))

	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error Posting alert to Discord \n%+v", err)
	}

	defer r.Body.Close()

	if r.StatusCode != 204 {
		contents, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("discord.SendAlerts: Error reading response body", err)
		}

		return fmt.Errorf(
			`discord.SendAlerts: Problem with Post, status code is not 204.
			StatusCode: %d, Message: %s, Request Body: %s, Response Body: %s`,
			r.StatusCode, r.Status, string(jsonDiscordMessage), string(contents))
	}

	return nil
}

func getDiscordChannel(
	discordChannelName string,
	configs config.Config) (config.DiscordChannel, error) {

	discordChannel, ok := configs.DiscordChannels[discordChannelName]
	if ok {
		return discordChannel, nil
	}

	err := fmt.Errorf(
		"discord.getDiscordChannel: The discordChannel %s could not be found",
		discordChannelName)
	return config.DiscordChannel{}, err
}

func createDiscordMessage(
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	discordChannel config.DiscordChannel,
	configs config.Config) (message WebhookParams, err error) {

	var contentBuilder strings.Builder

	handleMentions(alertmanagerBodyInfo, &contentBuilder, discordChannel, configs)

	var dashboardURL string
	if configs.DashboardLink.Enabled {
		dashboardURL = getDashboardURLFromGroup(alertmanagerBodyInfo, configs)
		if dashboardURL != "" {
			contentBuilder.WriteString(fmt.Sprintf("\n[%s](%s)",
				configs.DashboardLink.Text, dashboardURL))
		}
	}

	var generatorURL string
	if configs.GeneratorLink.Enabled {
		generatorURL = getGeneratorURLFromAlerts(alertmanagerBodyInfo)
		if generatorURL != "" {
			if configs.DashboardLink.Enabled && dashboardURL != "" {
				contentBuilder.WriteString("\n")
			}
			contentBuilder.WriteString(fmt.Sprintf("[%s](%s)",
				configs.GeneratorLink.Text, generatorURL))
		}
	}

	firingEmbeds, err := createDiscordMessageEmbeds(alertmanagerBodyInfo.FiringAlertsGroupedByName,
		"firing", configs)
	if err != nil {
		err = fmt.Errorf("discord.createDiscordMessage: Error creating firingEmbeds\n%+v", err)
		return WebhookParams{}, err
	}

	resolvedEmbeds, err := createDiscordMessageEmbeds(alertmanagerBodyInfo.ResolvedAlertsGroupedByName,
		"resolved", configs)
	if err != nil {
		err = fmt.Errorf("discord.createDiscordMessage: Error creating resolvedEmbeds %+v", err)
		return WebhookParams{}, err
	}

	embeds := append(firingEmbeds, resolvedEmbeds...)

	return WebhookParams{
		Content:   contentBuilder.String(),
		Embeds:    embeds,
		Username:  configs.Username,
		AvatarURL: configs.AvatarURL}, nil
}

func getDashboardURLFromGroup(alertmanagerBodyInfo alertmanager.MessageBodyInfo, configs config.Config) string {
	labelName := configs.DashboardLink.Label

	if url, ok := alertmanagerBodyInfo.GroupLabels[labelName]; ok && url != "" {
		return url
	}

	if url, ok := alertmanagerBodyInfo.CommonLabels[labelName]; ok && url != "" {
		return url
	}

	if url, ok := alertmanagerBodyInfo.CommonAnnotations[labelName]; ok && url != "" {
		return url
	}

	return ""
}

func getGeneratorURLFromAlerts(alertmanagerBodyInfo alertmanager.MessageBodyInfo) string {
	// Check firing alerts first
	for _, groupData := range alertmanagerBodyInfo.FiringAlertsGroupedByName {
		for _, alert := range groupData.Alerts {
			if alert.GeneratorURL != "" {
				return alert.GeneratorURL
			}
		}
	}

	// If not found in firing, check resolved alerts
	for _, groupData := range alertmanagerBodyInfo.ResolvedAlertsGroupedByName {
		for _, alert := range groupData.Alerts {
			if alert.GeneratorURL != "" {
				return alert.GeneratorURL
			}
		}
	}

	return ""
}

func handleMentions(
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	contentBuilder *strings.Builder,
	discordChannel config.DiscordChannel,
	configs config.Config) {

	var severitiesToMention []string

	// Channels can override global severitiesToMention
	if len(discordChannel.SeveritiesToMention) > 0 {
		severitiesToMention = discordChannel.SeveritiesToMention
	} else if len(configs.SeveritiesToMention) > 0 {
		severitiesToMention = configs.SeveritiesToMention
	}

	shouldMentionBySeverity := checkIfShouldMentionBySeverity(severitiesToMention, alertmanagerBodyInfo, configs)
	shouldMentionByFiringCount := checkIfShouldMentionByFiringCount(alertmanagerBodyInfo, configs)

	if shouldMentionBySeverity || shouldMentionByFiringCount {
		addRolesToEmbedContent(contentBuilder, discordChannel, configs)
	}

}

func checkIfShouldMentionBySeverity(
	severitiesToMention []string,
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	configs config.Config) bool {

	for _, severityToMention := range severitiesToMention {
		if alertmanagerBodyInfo.CountBySeverity[severityToMention] > 0 {
			return true
		}
	}

	return false
}

func checkIfShouldMentionByFiringCount(
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	configs config.Config) bool {

	if configs.FiringCountToMention > 0 {
		if alertmanagerBodyInfo.FiringCount >= configs.FiringCountToMention {
			return true
		}
	}

	return false

}

func addRolesToEmbedContent(
	contentBuilder *strings.Builder,
	discordChannel config.DiscordChannel,
	configs config.Config) {

	// Channels can override rolesToMention
	if len(discordChannel.RolesToMention) > 0 {
		contentBuilder.WriteString("    " + strings.Join(discordChannel.RolesToMention, " "))
	} else {
		contentBuilder.WriteString("    " + strings.Join(configs.RolesToMention, " "))
	}

}

func createDiscordMessageEmbeds(
	alertsGroupedByName alertmanager.AlertsGroupedByLabel,
	status string,
	configs config.Config) ([]MessageEmbed, error) {

	embedQueue := []EmbedQueueItem{}

	for _, groupData := range alertsGroupedByName {
		embed := MessageEmbed{}

		// Get title using correct Telegram template logic
		title := getAlertTitle(groupData.Alerts, groupData.GroupLabels)
		embed.Title = fmt.Sprintf("%s\n", title)

		description := ""
		for _, alert := range groupData.Alerts {
			alertText := "```"

			if configs.TimeDisplay.Enabled {
				alertText += "🔔\n"
			}

			if alert.Annotations["description"] == "" {
				alertText += "No description provided\n"
			} else {
				alertText += strings.TrimSuffix(strings.TrimSuffix(alert.Annotations["description"], "\n"), "\n") + "\n"
			}

			if configs.TimeDisplay.Enabled {
				timeInfo := formatAlertTimeInfo(alert, status, configs)
				if timeInfo != "" {
					alertText += "\n" + timeInfo
				}
			}

			alertText += "```"

			description += alertText
		}

		if len(description) > 0 {
			description = strings.TrimSuffix(description, "\n\n")
		}

		priority, err := handleEmbedAppearance(&embed, status, groupData.Alerts[0], configs)
		if err != nil {
			err = fmt.Errorf(
				`discord.createDiscordMessageEmbeds:
				Couldn't handle embed appearance for embed %+v and alert %+v: \n%+v`,
				embed, groupData.Alerts[0], err)
			return []MessageEmbed{}, err
		}

		title = "### " + embed.Title + "\n"
		embed.Title = ""

		embed.Description = title + description

		embedQueueItem := EmbedQueueItem{
			Embed:    embed,
			Priority: priority,
		}

		embedQueue = append(embedQueue, embedQueueItem)
	}

	sort.Slice(embedQueue[:], func(i, j int) bool {
		return embedQueue[i].Priority > embedQueue[j].Priority
	})

	embedsOrderedByPriority := []MessageEmbed{}

	for _, embedQueueItem := range embedQueue {
		embedsOrderedByPriority = append(embedsOrderedByPriority, embedQueueItem.Embed)
	}

	return embedsOrderedByPriority, nil
}

// getAlertTitle extracts title from group data following Telegram template logic
func getAlertTitle(alerts []alertmanager.Alert, groupLabels map[string]string) string {
	// 1. First try GroupLabels.summary
	if summary, ok := groupLabels["summary"]; ok && summary != "" {
		return summary
	}

	// 2. Search in alerts annotations
	for _, alert := range alerts {
		if summary, ok := alert.Annotations["summary"]; ok && summary != "" {
			return summary
		}
	}

	// 3. Try GroupLabels.alertname
	if alertname, ok := groupLabels["alertname"]; ok && alertname != "" {
		return alertname
	}

	// 4. Use alertname from first alert
	if len(alerts) > 0 {
		if alertname, ok := alerts[0].Labels["alertname"]; ok && alertname != "" {
			return alertname
		}
	}

	return "Unknown Alert"
}

func handleEmbedAppearance(
	embed *MessageEmbed, status string,
	alert alertmanager.Alert,
	configs config.Config) (priority int, err error) {

	if status == "resolved" {
		embed.Color = configs.Status["resolved"].Color
		embed.Title = fmt.Sprintf("\n%s %s", configs.Status["resolved"].Emoji, embed.Title)
		return 0, nil
	} else if status == "firing" {
		switch configs.MessageType {
		case "status":
			embed.Color = configs.Status["firing"].Color
			embed.Title = fmt.Sprintf("\n%s %s", configs.Status["firing"].Emoji, embed.Title)
			return 0, nil
		case "severity":
			severityAppearance := handleEmbedSeverity(embed, alert, configs)
			return severityAppearance.Priority, nil
		default:
			return 0, fmt.Errorf(
				"discord.handleEmbedAppearance: No matching message type for %s",
				configs.MessageType)
		}
	}

	return 0, nil
}

func handleEmbedSeverity(embed *MessageEmbed, alert alertmanager.Alert, configs config.Config) config.SeverityAppearance {
	severity, ok := alert.Labels[configs.Severity.Label]
	var SeverityAppearance config.SeverityAppearance
	if ok {
		SeverityAppearance, ok = configs.Severity.Values[severity]
		if !ok {
			SeverityAppearance = configs.Severity.Values["unknown"]
		}
		embed.Title = fmt.Sprintf("\n%s %s", SeverityAppearance.Emoji, embed.Title)
		embed.Color = SeverityAppearance.Color
	}
	return SeverityAppearance
}

func formatAlertTimeInfo(alert alertmanager.Alert, status string, configs config.Config) string {
	if !configs.TimeDisplay.Enabled {
		return ""
	}

	var timeInfo strings.Builder

	startsAt, err := time.Parse(time.RFC3339, alert.StartsAt)
	if err != nil {
		log.Printf("ERROR: Failed to parse StartsAt time: %v", err)
		return ""
	}

	localStartsAt := startsAt.Format("02.01.2006 15:04:05 MST")

	timeInfo.WriteString(fmt.Sprintf("🕑\n%s %s", configs.TimeDisplay.StartsAtText, localStartsAt))

	if status == "resolved" && alert.EndsAt != "" {
		endsAt, err := time.Parse(time.RFC3339, alert.EndsAt)
		if err != nil {
			log.Printf("ERROR: Failed to parse EndsAt time: %v", err)
		} else {
			localEndsAt := endsAt.Format("02.01.2006 15:04:05 MST")
			duration := endsAt.Sub(startsAt)

			// Format duration
			durationStr := formatDuration(duration)

			timeInfo.WriteString(fmt.Sprintf("\n%s %s", configs.TimeDisplay.EndsAtText, localEndsAt))
			timeInfo.WriteString(fmt.Sprintf("\n%s %s", configs.TimeDisplay.DurationText, durationStr))
		}
	}

	return timeInfo.String()
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	var parts []string

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, " ")
}
