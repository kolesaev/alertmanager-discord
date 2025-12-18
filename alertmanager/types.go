package alertmanager

// Alert represents the alert object in Alertmanager's webhook body
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	GeneratorURL string            `json:"generatorURL"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	Fingerprint  string            `json:"fingerprint"`
}

// MessageBody represents the fields available in Alertmanager's webhook
type MessageBody struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TrucatedAlerts    int32             `json:"truncatedAlerts"`
}

// MessageBodyInfo is a type with the necessary info to perform the checks
// and construct other objects, e.g. Discord's WebhookParams
type MessageBodyInfo struct {
	FiringCount, ResolvedCount  int
	CountBySeverity             map[string]int
	FiringAlertsGroupedByName   AlertsGroupedByLabel
	ResolvedAlertsGroupedByName AlertsGroupedByLabel
	GroupLabels                 map[string]string
	CommonLabels                map[string]string
	CommonAnnotations           map[string]string
	ExternalURL                 string
}

// AlertsGroupedByLabel is just a wrapper for the common case of grouping
// alerts by label, such as "alertname"
type AlertsGroupedByLabel map[string]struct {
	Alerts      []Alert
	GroupLabels map[string]string
}
