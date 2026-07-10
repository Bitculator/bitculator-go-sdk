package bitculator

import (
	"context"
	"encoding/json"
)

// Editorial covers editorial content: articles, videos, and insights.
type Editorial struct{ c *httpClient }

// CoinVideos returns videos for a coin (paginated). Params: page, per_page, type, search.
func (r *Editorial) CoinVideos(ctx context.Context, slug string, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/coins/"+enc(slug)+"/videos", params)
}

// CoinInsights returns the insight timeline for a coin. Params: locale, offset, limit.
func (r *Editorial) CoinInsights(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/insights", params)
}

// Articles lists articles (paginated). Params: page, per_page, locale, tag, coin,
// exchange, wallet, search.
func (r *Editorial) Articles(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/articles", params)
}

// Article returns a single article by slug. Params: locale.
func (r *Editorial) Article(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/articles/"+enc(slug), params)
}

// SubmitArticleFeedback submits helpful/unhelpful feedback on an article.
func (r *Editorial) SubmitArticleFeedback(ctx context.Context, slug string, helpful bool) (json.RawMessage, error) {
	return r.c.post(ctx, "/api/v1/articles/"+enc(slug)+"/feedback", map[string]any{"helpful": helpful})
}

// Video returns a single video by id.
func (r *Editorial) Video(ctx context.Context, id any) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/videos/"+enc(id), nil)
}

// Insights lists insights (paginated). Params: page, per_page, locale, type,
// coin, search, sort.
func (r *Editorial) Insights(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/insights", params)
}

// Insight returns a single insight by id. Params: locale.
func (r *Editorial) Insight(ctx context.Context, id any, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/insights/"+enc(id), params)
}

// AlarmBody is the body for creating an alarm.
type AlarmBody struct {
	Name         string  `json:"name"`
	Coin         string  `json:"coin"`
	Metric       string  `json:"metric"`
	Direction    string  `json:"direction"`
	Target       float64 `json:"target"`
	Notification string  `json:"notification"`
}

// Alarms covers price alarms owned by the API key.
type Alarms struct{ c *httpClient }

// List returns your alarms (paginated). Params: page, per_page, status,
// direction, notification.
func (r *Alarms) List(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/alarms", params)
}

// Create creates an alarm.
func (r *Alarms) Create(ctx context.Context, alarm AlarmBody) (json.RawMessage, error) {
	return r.c.post(ctx, "/api/v1/alarms", alarm)
}

// Delete deletes an alarm by id.
func (r *Alarms) Delete(ctx context.Context, id any) (json.RawMessage, error) {
	return r.c.delete(ctx, "/api/v1/alarms/"+enc(id))
}

// WebhookBody is the body for registering a webhook endpoint.
type WebhookBody struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// Webhooks covers webhook endpoints and their delivery log.
type Webhooks struct{ c *httpClient }

// List returns your webhook endpoints.
func (r *Webhooks) List(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/webhooks", nil)
}

// Create registers a webhook endpoint for a set of events.
func (r *Webhooks) Create(ctx context.Context, webhook WebhookBody) (json.RawMessage, error) {
	return r.c.post(ctx, "/api/v1/webhooks", webhook)
}

// Delete deletes a webhook endpoint by id.
func (r *Webhooks) Delete(ctx context.Context, id any) (json.RawMessage, error) {
	return r.c.delete(ctx, "/api/v1/webhooks/"+enc(id))
}

// Test sends a test event to a webhook endpoint.
func (r *Webhooks) Test(ctx context.Context, id any) (json.RawMessage, error) {
	return r.c.post(ctx, "/api/v1/webhooks/"+enc(id)+"/test", nil)
}

// Deliveries returns the delivery log for a webhook endpoint (paginated). Params:
// page, per_page.
func (r *Webhooks) Deliveries(ctx context.Context, id any, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/webhooks/"+enc(id)+"/deliveries", params)
}

// Meta covers service metadata: spec, health, and key usage.
type Meta struct{ c *httpClient }

// OpenAPI returns the raw OpenAPI specification for the API.
func (r *Meta) OpenAPI(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/openapi.json", nil)
}

// Ping is a health check.
func (r *Meta) Ping(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/ping", nil)
}

// Usage returns the current key usage and quota.
func (r *Meta) Usage(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/usage", nil)
}
