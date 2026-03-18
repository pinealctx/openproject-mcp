package openproject

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
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
	// Filter by read status: "" = all, "f" = unread, "t" = read
	ReadIAN  string
	PageSize int
	Offset   int
}

// ListNotifications returns the current user's notifications.
func (c *Client) ListNotifications(ctx context.Context, opts *ListNotificationsOptions) (*NotificationList, error) {
	params := url.Values{}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(opts.PageSize))
	}
	if opts.ReadIAN != "" {
		// OpenProject API expects "f" for false (unread), "t" for true (read)
		filterJSON := fmt.Sprintf(`[{"readIAN":{"operator":"=","values":["%s"]}}]`, opts.ReadIAN)
		params.Set("filters", filterJSON)
	}

	path := "/notifications"
	if len(params) > 0 {
		path += "?" + params.Encode()
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
