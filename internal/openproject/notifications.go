package openproject

import (
	"context"
	"fmt"
	"time"
)

// NotificationReason represents why a notification was created.
type NotificationReason struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Href  string `json:"href"`
}

// NotificationLinks represents _links in a notification.
type NotificationLinks struct {
	Actor    *Link `json:"actor,omitempty"`
	Project  *Link `json:"project,omitempty"`
	Resource *Link `json:"resource,omitempty"`
	Activity *Link `json:"activity,omitempty"`
}

// Notification represents a single OpenProject notification.
type Notification struct {
	ID        int                  `json:"id"`
	Reason    string               `json:"reason"`
	ReadIAN   *bool                `json:"readIAN"`
	CreatedAt *time.Time           `json:"createdAt"`
	UpdatedAt *time.Time           `json:"updatedAt"`
	Details   []NotificationReason `json:"details,omitempty"`
	Links     *NotificationLinks   `json:"_links,omitempty"`
}

// NotificationList represents a paginated list of notifications.
type NotificationList struct {
	Embedded struct {
		Elements []Notification `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// ListNotificationsOptions contains options for listing notifications.
type ListNotificationsOptions struct {
	// Filter by read status: "" = all, "true" = read, "false" = unread
	ReadIAN  string
	PageSize int
	Offset   int
}

// ListNotifications returns the current user's notifications.
func (c *Client) ListNotifications(ctx context.Context, opts *ListNotificationsOptions) (*NotificationList, error) {
	path := "/notifications"
	params := buildQuery(map[string]string{
		"pageSize": intParam(opts.PageSize),
		"offset":   intParam(opts.Offset),
	})
	if opts.ReadIAN != "" {
		params += fmt.Sprintf(`&filters=[{"readIAN":{"operator":"=","values":[%q]}}]`, opts.ReadIAN)
	}
	if params != "" {
		path += "?" + params[1:] // strip leading "&"
	}

	var list NotificationList
	if err := c.Get(ctx, path, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// MarkNotificationRead marks a single notification as read.
func (c *Client) MarkNotificationRead(ctx context.Context, id int) error {
	return c.Post(ctx, fmt.Sprintf("/notifications/%d/read_ian", id), nil, nil)
}

// MarkAllNotificationsRead marks all notifications as read.
func (c *Client) MarkAllNotificationsRead(ctx context.Context) error {
	return c.Post(ctx, "/notifications/read_ian", nil, nil)
}

// buildQuery is a small helper to build a query string from a string map,
// skipping empty values. Returns "" or "&k=v&k=v..." (leading &, not ?).
func buildQuery(params map[string]string) string {
	result := ""
	for k, v := range params {
		if v != "" && v != "0" {
			result += fmt.Sprintf("&%s=%s", k, v)
		}
	}
	return result
}

// intParam converts an int to a string param, returning "" for zero.
func intParam(v int) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%d", v)
}
