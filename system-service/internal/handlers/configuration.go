package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bowens/kabletown/shared/response"
)

// GetConfiguration handles GET /System/Configuration.
func (h *Handler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"LogFileRetentionDays":              3,
		"IsStartupWizardCompleted":          true,
		"CachePath":                         "/cache",
		"AutoRunWebApp":                     false,
		"EnableUPnP":                        false,
		"EnableMetrics":                     false,
		"PublicPort":                        8096,
		"UPnPCreateHttpPortMap":             false,
		"PublicHttpsPort":                   8920,
		"HttpServerPortNumber":              8096,
		"HttpsPortNumber":                   8920,
		"EnableHttps":                       false,
		"RequireHttps":                      false,
		"CertificatePath":                   "",
		"CertificatePassword":               "",
		"BaseUrl":                           "/",
		"EnableNormalizedItemByNameIds":      false,
		"DisableLiveTvChannelUserDataName":   true,
		"MetadataPath":                      "/config/metadata",
		"MetadataNetworkPath":               "",
		"PreferredMetadataLanguage":         "en",
		"MetadataCountryCode":               "US",
		"SortReplaceCharacters":             []string{".", "+", "%"},
		"SortRemoveCharacters":              []string{",", "&", "-", "{", "}", "'"},
		"SortRemoveWords":                   []string{"the", "a", "an"},
		"MinResumePct":                      5,
		"MaxResumePct":                      90,
		"MinResumeDurationSeconds":          300,
		"MinAudiobookResume":                5,
		"MaxAudiobookResume":                5,
		"InactiveSessionThreshold":          0,
		"LibraryMonitorDelay":               60,
		"LibraryUpdateDuration":             30,
		"ImageSavingConvention":             "Legacy",
		"MetadataOptions":                   []interface{}{},
		"SkipDeserializationForBasicTypes":   false,
		"ServerName":                         h.serverName,
		"UICulture":                          "en-US",
		"SaveMetadataHidden":                 false,
		"ContentTypes":                       []interface{}{},
		"RemoteClientBitrateLimit":           0,
		"EnableFolderView":                   false,
		"EnableGroupingIntoCollections":      false,
		"DisplaySpecialsWithinSeasons":       true,
		"CodecsUsed":                         []interface{}{},
		"PluginRepositories":                 []interface{}{},
		"EnableExternalContentInSuggestions": true,
		"ImageExtractionTimeoutMs":           0,
		"PathSubstitutions":                  []interface{}{},
		"EnableSlowResponseWarning":          true,
		"SlowResponseThresholdMs":            500,
		"CorsHosts":                          []string{"*"},
		"ActivityLogRetentionDays":           30,
		"LibraryScanFanoutConcurrency":       0,
		"LibraryMetadataRefreshConcurrency":  0,
		"RemoveOldPlugins":                   false,
		"AllowClientLogUpload":               true,
	})
}

// UpdateConfiguration handles POST /System/Configuration.
func (h *Handler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// GetNamedConfiguration handles GET /System/Configuration/{key}.
func (h *Handler) GetNamedConfiguration(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "key")
	response.JSON(w, http.StatusOK, map[string]interface{}{})
}

// UpdateNamedConfiguration handles POST /System/Configuration/{key}.
func (h *Handler) UpdateNamedConfiguration(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
