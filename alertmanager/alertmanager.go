package alertmanager

import (
	"github.com/kolesaev/alertmanager-discord/config"
)

// ExtractBodyInfo Extracts the necessary info to perform the checks and construct the Discord
// message body more easily
func ExtractBodyInfo(alertmanagerBody MessageBody, config config.Config) MessageBodyInfo {

	alerts := alertmanagerBody.Alerts

	firingCount := 0
	resolvedCount := 0
	countBySeverity := make(map[string]int)
	firingAlertsGroupedByName := AlertsGroupedByLabel{}
	resolvedAlertsGroupedByName := AlertsGroupedByLabel{}

	alertName := ""

	for _, alert := range alerts {
		alertName = alert.Labels["alertname"]
		status := alert.Status

		severityValue, ok := alert.Labels[config.Severity.Label]
		if ok {
			countBySeverity[severityValue]++
		} else {
			alert.Labels[config.Severity.Label] = "unknown"
			countBySeverity["unknown"]++
		}

		if status == "firing" {
			firingCount++
			group := firingAlertsGroupedByName[alertName]
			group.Alerts = append(group.Alerts, alert)
			group.GroupLabels = alertmanagerBody.GroupLabels
			firingAlertsGroupedByName[alertName] = group
		} else if status == "resolved" {
			resolvedCount++
			group := resolvedAlertsGroupedByName[alertName]
			group.Alerts = append(group.Alerts, alert)
			group.GroupLabels = alertmanagerBody.GroupLabels
			resolvedAlertsGroupedByName[alertName] = group
		}
	}

	return MessageBodyInfo{
		FiringCount:                 firingCount,
		ResolvedCount:               resolvedCount,
		CountBySeverity:             countBySeverity,
		FiringAlertsGroupedByName:   firingAlertsGroupedByName,
		ResolvedAlertsGroupedByName: resolvedAlertsGroupedByName,
		GroupLabels:                 alertmanagerBody.GroupLabels,
		CommonLabels:                alertmanagerBody.CommonLabels,
		CommonAnnotations:           alertmanagerBody.CommonAnnotations,
		ExternalURL:                 alertmanagerBody.ExternalURL,
	}
}

// CheckIfHasOnlySeveritiesToIgnoreWhenAlone verifies if in the countBySeverity
// map there are only severities that should be ignored when alone. It first
// tries to use the array "SeveritiesToIgnoreWhenAlone" defined in the Discord
// Channel, then in the global config, and if there isn't any, returns false.
func CheckIfHasOnlySeveritiesToIgnoreWhenAlone(
	countBySeverity map[string]int,
	discordChannel config.DiscordChannel,
	configs config.Config) bool {

	var severitiesToIgnore []string

	if len(discordChannel.SeveritiesToIgnoreWhenAlone) > 0 {
		severitiesToIgnore = discordChannel.SeveritiesToIgnoreWhenAlone
	} else if len(configs.SeveritiesToIgnoreWhenAlone) > 0 {
		severitiesToIgnore = configs.SeveritiesToIgnoreWhenAlone
	} else {
		return false
	}

	for severity := range countBySeverity {
		if !contains(severitiesToIgnore, severity) {
			return false
		}
	}

	return true
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
