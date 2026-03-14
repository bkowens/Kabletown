// Package dto provides data transfer objects for Kabletown services
package dto

// UserPolicyDto contains user policy and permission settings
type UserPolicyDto struct {
	// Core
	Id                  string                `json:"id"`
	UserId              string                `json:"userId"`
	Name                string                `json:"name"`
	Role                string                `json:"role"`
	
	// Authentication & Access
	IsAdmin             bool                  `json:"isAdmin"`
	IsHidden            bool                  `json:"isHidden"`
	IsDisabled          bool                  `json:"isDisabled"`
	IsLockedOut         bool                  `json:"isLockedOut"`
	BlockAccessToMetadataChapters bool       `json:"blockAccessToMetadataChapters"`
	EnableContentDeletion bool                `json:"enableContentDeletion"`
	EnableContentDeletionFromFolders []string `json:"enableContentDeletionFromFolders"`
	EnableCollectionManagement bool            `json:"enableCollectionManagement"`
	EnablePublicUser      bool                `json:"enablePublicUser"`
	EnableAudioPlayback   bool                `json:"enableAudioPlayback"`
	EnableVideoPlayback   bool                `json:"enableVideoPlayback"`
	
	// Authentication
	AuthenticationProviderId string   `json:"authenticationProviderId"`
	PasswordResetProviderId  string   `json:"passwordResetProviderId"`
	
	// Password
	ResetPasswordLink     *string `json:"resetPasswordLink,omitempty"`
	EnableUserPasswordReset bool  `json:"enableUserPasswordReset"`
	
	// Permissions - map of permission kind (int) to boolean (enabled/disabled)
	// Permission kinds are defined in shared/types/permissions.go
	Permissions map[int]bool `json:"permissions"`
	
	// Allowed Audio/Video Channels
	MaxActiveStreams    int                   `json:"maxActiveStreams"`
	
	// Parental Controls
	ParentalRating      *int                  `json:"parentalRating,omitempty"`
	MaxParentalRating   *int                  `json:"maxParentalRating,omitempty"`
	
	// Library Access
	EnableAllLibraries   bool                  `json:"enableAllLibraries"`
	AllowedFolders       []string              `json:"allowedFolders"`  // Folder GUIDs
	BlockedFolders       []string              `json:"blockedFolders"`
	BlockedScreenCasting bool                  `json:"blockedScreenCasting"`
	
	// Content Restrictions
	BlockedMediaTypes    []string `json:"blockedMediaTypes"`
	BlockedIndexes       []string `json:"blockedIndexes"`
	
	// SyncPlay
	CreateGroup         bool     `json:"createGroup"`
	JoinAnyGroup        bool     `json:"joinAnyGroup"`
	
	// Device Limits
	MaxActiveSessions   int      `json:"maxActiveSessions"`
	
	// Timestamps
	CreatedAt           string   `json:"createdAt,omitempty"`
	UpdatedAt           string   `json:"updatedAt,omitempty"`
}

// PermissionKind defines the enum values for permission types
// These should match the database Permissions table
type PermissionKind int

const (
	// Standard Jellyfin permission kinds
	PermissionKeyAllowPlaybackControl        PermissionKind = 1
	PermissionKeyAllowContentDeletion       PermissionKind = 2
	PermissionKeyAllowCollectionManagement  PermissionKind = 3
	PermissionKeyAllowPublicUser            PermissionKind = 4
	PermissionKeyAllowAudioPlayback         PermissionKind = 5
	PermissionKeyAllowVideoPlayback         PermissionKind = 6
	PermissionKeyAllowLiveTvManagement      PermissionKind = 7
	PermissionKeyAllowPublicChannels        PermissionKind = 8
	PermissionKeyAllowRemoteAccess          PermissionKind = 9
	PermissionKeyAllowSyncPlayback          PermissionKind = 10
)

// PermissionInfo contains a single permission mapping
// Used for API request/response
type PermissionInfo struct {
	Kind    int      `json:"kind"`     // Permission kind ID
	Value   bool     `json:"value"`    // Permission enabled/disabled
	Name    string   `json:"name"`     // Human-readable name
	Default bool     `json:"default"`  // Default value for new users
}

// PermissionsMapFromList converts a list to a map for efficient lookups
func PermissionsMapFromList(permissions []PermissionInfo) map[int]bool {
	result := make(map[int]bool, len(permissions))
	for _, p := range permissions {
		result[p.Kind] = p.Value
	}
	return result
}

// PermissionsListFromMap converts a map to a list for API serialization
func PermissionsListFromMap(permissions map[int]bool) []PermissionInfo {
	result := make([]PermissionInfo, 0, len(permissions))
	for kind, value := range permissions {
		result = append(result, PermissionInfo{
			Kind:  kind,
			Value: value,
		})
	}
	return result
}

// HasPermission checks if a user has a specific permission
func (p *UserPolicyDto) HasPermission(kind int) bool {
	if p.Permissions == nil {
		return false
	}
	value, exists := p.Permissions[kind]
	return exists && value
}

// SetPermission sets a permission value
func (p *UserPolicyDto) SetPermission(kind int, value bool) {
	if p.Permissions == nil {
		// Initialize map on first use
		p.Permissions = make(map[int]bool)
	}
	p.Permissions[kind] = value
}

// GetAllowedFolders returns the list of folders a user can access
func (p *UserPolicyDto) GetAllowedFolders() []string {
	if p.EnableAllLibraries {
		return []string{} // Empty list means all
	}
	return p.AllowedFolders
}

// HasFolderAccess checks if user has access to a specific folder
func (p *UserPolicyDto) HasFolderAccess(folderId string) bool {
	if p.EnableAllLibraries {
		return true
	}
	for _, allowedId := range p.AllowedFolders {
		if allowedId == folderId {
			return true
		}
	}
	return false
}

// IsBlockedFromFolder checks if a folder is explicitly blocked
func (p *UserPolicyDto) IsBlockedFromFolder(folderId string) bool {
	for _, blockedId := range p.BlockedFolders {
		if blockedId == folderId {
			return true
		}
	}
	return false
}

// MessageDto is a generic message container for API responses
type MessageDto struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// MetadataRefreshTask represents a metadata refresh job
type MetadataRefreshTask struct {
	Id       string    `json:"id,omitempty"`
	ItemId   string    `json:"itemId"`
	UserId   string    `json:"userId,omitempty"`
	Overwrite bool     `json:"overwrite"`
	Recursive bool     `json:"recursive"`
	Status   string    `json:"status"` // Queued, Running, Completed, Failed, Cancelled
	CreatedAt string   `json:"createdAt"`
	StartedAt *string   `json:"startedAt,omitempty"`
	CompletedAt *string `json:"completedAt,omitempty"`
	Error     *string   `json:"error,omitempty"`
}

// ScheduledTask represents a scheduled system task
type ScheduledTaskDto struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	LastRunTime   string `json:"lastRunTime,omitempty"`
	LastDuration  int64  `json:"lastDuration,omitempty"` // in milliseconds
	NextRunTime   string `json:"nextRunTime,omitempty"`
	Status        string `json:"status"` // Running, Idle, Failed, Canceled
	State         string `json:"state"`  // Idle, Running, Waiting, WaitingForDependencies, PostTaskAction, Finished
	TotalRunningSeconds int `json:"totalRunningSeconds"`
	TotalFailedRuns int `json:"totalFailedRuns"`
	ScheduleString string `json:"scheduleString"`
}

// ImageInfoDto contains image metadata for library items
type ImageInfoDto struct {
	Url         string            `json:"url"`
	Provider    string            `json:"provider"`
	ProviderId  string            `json:"providerId"`
	Height      int               `json:"height,omitempty"`
	Width       int               `json:"width,omitempty"`
	AspectRatio *float64          `json:"aspectRatio,omitempty"`
	SourceType  string            `json:"sourceType"` // Local, Remote, Embedded
	ImageType   string            `json:"imageType"`  // Primary, Banner, Disc, Logo, Backdrop, Thumb, Menu
	Language    *string           `json:"language,omitempty"`
	Rating      *int              `json:"rating,omitempty"`
	DateAdded   string            `json:"dateAdded"`


}
