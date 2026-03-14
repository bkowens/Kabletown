package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// GetLocalizationOptions handles GET /Localization/Options.
func (h *Handler) GetLocalizationOptions(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []map[string]string{
		{"Name": "Arabic", "Value": "ar"},
		{"Name": "Chinese Simplified", "Value": "zh-CN"},
		{"Name": "Chinese Traditional", "Value": "zh-TW"},
		{"Name": "Czech", "Value": "cs"},
		{"Name": "Danish", "Value": "da"},
		{"Name": "Dutch", "Value": "nl"},
		{"Name": "English (United Kingdom)", "Value": "en-GB"},
		{"Name": "English (United States)", "Value": "en-US"},
		{"Name": "Finnish", "Value": "fi"},
		{"Name": "French", "Value": "fr"},
		{"Name": "German", "Value": "de"},
		{"Name": "Hebrew", "Value": "he"},
		{"Name": "Hungarian", "Value": "hu"},
		{"Name": "Italian", "Value": "it"},
		{"Name": "Japanese", "Value": "ja"},
		{"Name": "Korean", "Value": "ko"},
		{"Name": "Norwegian Bokmål", "Value": "nb"},
		{"Name": "Polish", "Value": "pl"},
		{"Name": "Portuguese (Brazil)", "Value": "pt-BR"},
		{"Name": "Portuguese (Portugal)", "Value": "pt-PT"},
		{"Name": "Romanian", "Value": "ro"},
		{"Name": "Russian", "Value": "ru"},
		{"Name": "Spanish", "Value": "es"},
		{"Name": "Spanish (Latin America)", "Value": "es-419"},
		{"Name": "Swedish", "Value": "sv"},
		{"Name": "Turkish", "Value": "tr"},
		{"Name": "Ukrainian", "Value": "uk"},
	})
}

// GetCountries handles GET /Localization/Countries.
func (h *Handler) GetCountries(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []map[string]string{
		{"Name": "United States", "DisplayName": "United States", "TwoLetterISORegionName": "US", "ThreeLetterISORegionName": "USA"},
		{"Name": "United Kingdom", "DisplayName": "United Kingdom", "TwoLetterISORegionName": "GB", "ThreeLetterISORegionName": "GBR"},
		{"Name": "Canada", "DisplayName": "Canada", "TwoLetterISORegionName": "CA", "ThreeLetterISORegionName": "CAN"},
		{"Name": "Australia", "DisplayName": "Australia", "TwoLetterISORegionName": "AU", "ThreeLetterISORegionName": "AUS"},
		{"Name": "Germany", "DisplayName": "Germany", "TwoLetterISORegionName": "DE", "ThreeLetterISORegionName": "DEU"},
		{"Name": "France", "DisplayName": "France", "TwoLetterISORegionName": "FR", "ThreeLetterISORegionName": "FRA"},
		{"Name": "Japan", "DisplayName": "Japan", "TwoLetterISORegionName": "JP", "ThreeLetterISORegionName": "JPN"},
	})
}

// GetCultures handles GET /Localization/Cultures.
func (h *Handler) GetCultures(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []map[string]interface{}{
		{"Name": "English", "DisplayName": "English", "TwoLetterISOLanguageName": "en", "ThreeLetterISOLanguageName": "eng", "ThreeLetterISOLanguageNameList": []string{"eng"}},
		{"Name": "French", "DisplayName": "French", "TwoLetterISOLanguageName": "fr", "ThreeLetterISOLanguageName": "fre", "ThreeLetterISOLanguageNameList": []string{"fre", "fra"}},
		{"Name": "German", "DisplayName": "German", "TwoLetterISOLanguageName": "de", "ThreeLetterISOLanguageName": "ger", "ThreeLetterISOLanguageNameList": []string{"ger", "deu"}},
		{"Name": "Spanish", "DisplayName": "Spanish", "TwoLetterISOLanguageName": "es", "ThreeLetterISOLanguageName": "spa", "ThreeLetterISOLanguageNameList": []string{"spa"}},
		{"Name": "Japanese", "DisplayName": "Japanese", "TwoLetterISOLanguageName": "ja", "ThreeLetterISOLanguageName": "jpn", "ThreeLetterISOLanguageNameList": []string{"jpn"}},
		{"Name": "Chinese", "DisplayName": "Chinese", "TwoLetterISOLanguageName": "zh", "ThreeLetterISOLanguageName": "chi", "ThreeLetterISOLanguageNameList": []string{"chi", "zho"}},
	})
}

// GetParentalRatings handles GET /Localization/ParentalRatings.
func (h *Handler) GetParentalRatings(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []map[string]interface{}{
		{"Name": "NR", "Value": -1},
		{"Name": "G", "Value": 1},
		{"Name": "PG", "Value": 5},
		{"Name": "PG-13", "Value": 7},
		{"Name": "R", "Value": 9},
		{"Name": "NC-17", "Value": 10},
	})
}
