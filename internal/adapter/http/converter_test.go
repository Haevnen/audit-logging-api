package handler_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	h "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/pkg/utils"
)

func TestToEntityAction(t *testing.T) {
	assert.Equal(t, entitylog.ActionCreate, h.ToEntityAction(api_service.CREATE))
	assert.Equal(t, entitylog.ActionUpdate, h.ToEntityAction(api_service.UPDATE))
	assert.Equal(t, entitylog.ActionDelete, h.ToEntityAction(api_service.DELETE))
	assert.Equal(t, entitylog.ActionView, h.ToEntityAction(api_service.VIEW))
	assert.Equal(t, entitylog.ActionType(""), h.ToEntityAction("INVALID"))
}

func TestToEntitySeverity(t *testing.T) {
	assert.Equal(t, entitylog.SeverityCritical, h.ToEntitySeverity(api_service.CRITICAL))
	assert.Equal(t, entitylog.SeverityError, h.ToEntitySeverity(api_service.ERROR))
	assert.Equal(t, entitylog.SeverityWarning, h.ToEntitySeverity(api_service.WARNING))
	assert.Equal(t, entitylog.SeverityInfo, h.ToEntitySeverity(api_service.INFO))
	assert.Equal(t, entitylog.Severity(""), h.ToEntitySeverity("INVALID"))
}

func TestMarshallDataAndJSONToMap(t *testing.T) {
	// nil input
	j, err := h.MarshallData(nil)
	assert.NoError(t, err)
	assert.Nil(t, j)

	// normal map
	m := map[string]interface{}{"foo": "bar"}
	j, err = h.MarshallData(&m)
	assert.NoError(t, err)
	assert.NotNil(t, j)

	back, err := h.JSONToMap(j)
	assert.NoError(t, err)
	assert.Equal(t, "bar", (*back)["foo"])

	// empty JSON
	empty := datatypes.JSON([]byte{})
	back, err = h.JSONToMap(&empty)
	assert.NoError(t, err)
	assert.Nil(t, back)
}

func TestToLogStatsResponse(t *testing.T) {
	now := time.Now()
	stats := []entitylog.LogStats{
		{
			Day:              now,
			Total:            10,
			ActionCreate:     1,
			ActionUpdate:     2,
			ActionDelete:     3,
			ActionView:       4,
			SeverityInfo:     5,
			SeverityError:    6,
			SeverityWarning:  7,
			SeverityCritical: 8,
		},
	}
	resp := h.ToLogStatsResponse(stats)
	assert.Len(t, resp, 1)
	assert.Equal(t, now.Format("2006-01-02"), resp[0].Day)
	assert.Equal(t, int64(10), resp[0].Total)
	assert.Equal(t, int64(1), resp[0].CREATE)
	assert.Equal(t, int64(8), resp[0].CRITICAL)
}

func TestToSingleLogResponse(t *testing.T) {
	now := time.Now()

	meta := map[string]interface{}{"a": "b"}
	j, _ := json.Marshal(meta)
	dj := datatypes.JSON(j)

	log := entitylog.Log{
		ID:             "id1",
		UserID:         "u1",
		TenantID:       "t1",
		Action:         entitylog.ActionCreate,
		Severity:       entitylog.SeverityInfo,
		EventTimestamp: now,
		Message:        "msg",
		SessionID:      utils.Ptr("s1"),
		Resource:       utils.Ptr("res"),
		ResourceID:     utils.Ptr("rid"),
		IPAddress:      utils.Ptr("127.0.0.1"),
		UserAgent:      utils.Ptr("ua"),
		BeforeState:    &dj,
		AfterState:     &dj,
		Metadata:       &dj,
	}

	resp, err := h.ToSingleLogResponse(log)
	assert.NoError(t, err)
	assert.Equal(t, "id1", resp.Id)
	assert.Equal(t, "u1", resp.UserId)
	assert.Equal(t, "t1", resp.TenantId)
	assert.Equal(t, "msg", resp.Message)
	assert.Equal(t, "s1", *resp.SessionId)        // ✅ dereference
	assert.Equal(t, "res", *resp.Resource)        // ✅ dereference
	assert.Equal(t, "rid", *resp.ResourceId)      // ✅ dereference
	assert.Equal(t, "127.0.0.1", *resp.IpAddress) // ✅ dereference
	assert.Equal(t, "ua", *resp.UserAgent)        // ✅ dereference
	assert.Equal(t, "b", (*resp.Metadata)["a"])
}
