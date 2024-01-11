package projects

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"

	"github.com/gorilla/context"
)

func WebhookExtractorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractor_id, err := helpers.GetIntParam("extractor_id", w, r)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid extractor ID",
			})
			return
		}

		webhook_id, err := helpers.GetIntParam("webhook_id", w, r)
		var extractor db.WebhookExtractor
		extractor, err = helpers.Store(r).GetWebhookExtractor(extractor_id, webhook_id)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "extractor", extractor)
		next.ServeHTTP(w, r)
	})
}

func GetWebhookExtractor(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)

	helpers.WriteJSON(w, http.StatusOK, extractor)
}

func GetWebhookExtractors(w http.ResponseWriter, r *http.Request) {
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)
	extractors, err := helpers.Store(r).GetWebhookExtractors(webhook_id, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, extractors)
}

func AddWebhookExtractor(w http.ResponseWriter, r *http.Request) {
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)
	var extractor db.WebhookExtractor

	if !helpers.Bind(w, r, &extractor) {
		return
	}

	if extractor.WebhookID != webhook_id {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Webhook ID in body and URL must be the same",
		})
		return
	}

	if err := extractor.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	_, err = helpers.Store(r).CreateWebhookExtractor(extractor)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateWebhookExtractor(w http.ResponseWriter, r *http.Request) {
	oldExtractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var extractor db.WebhookExtractor

	if !helpers.Bind(w, r, &extractor) {
		return
	}

	if extractor.ID != oldExtractor.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "WebhookExtractor ID in body and URL must be the same",
		})
		return
	}

	if extractor.WebhookID != oldExtractor.WebhookID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Webhook ID in body and URL must be the same",
		})
		return
	}

	err := helpers.Store(r).UpdateWebhookExtractor(extractor)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetWebhookExtractorRefs(w http.ResponseWriter, r *http.Request) {
	extractor_id, err := helpers.GetIntParam("extractor_id", w, r)

	log.Info(fmt.Sprintf("Extractor ID: %v", extractor_id))
	fmt.Println(fmt.Sprintf("Extractor ID: %v", extractor_id))
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Extractor ID",
		})
		return
	}

	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

	var extractor db.WebhookExtractor
	extractor, err = helpers.Store(r).GetWebhookExtractor(webhook_id, extractor_id)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	refs, err := helpers.Store(r).GetWebhookExtractorRefs(extractor.WebhookID, extractor.ID)
	log.Info(fmt.Sprintf("References found: %v", refs))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func DeleteWebhookExtractor(w http.ResponseWriter, r *http.Request) {
	extractor_id, err := helpers.GetIntParam("extractor_id", w, r)
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

	log.Info(fmt.Sprintf("Delete requested for: %v", extractor_id))
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Extractor ID",
		})
		return
	}

	var extractor db.WebhookExtractor
	extractor, err = helpers.Store(r).GetWebhookExtractor(webhook_id, extractor_id)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).DeleteWebhookExtractor(webhook_id, extractor.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Webhook Extractor failed to be deleted",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
