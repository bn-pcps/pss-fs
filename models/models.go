package models

import (
	"time"

	"github.com/google/uuid"
)

// TableName methods ensure GORM uses the correct table names

// PsPlans represents the ps_plans table
type PsPlans struct {
	ID       int     `json:"id" gorm:"primaryKey"`
	PolarId  *string `json:"polar_id" gorm:"column:polar_id;size:255"`
	PlanName string  `json:"plan_name" gorm:"column:plan_name;size:100;not null"`
	Quota    int64   `json:"quota" gorm:"not null"` // in MB
}

func (PsPlans) TableName() string {
	return "ps_plans"
}

// PsUsers represents the ps_users table
type PsUsers struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	GoogleId  string    `json:"google_id" gorm:"column:google_id;size:255;uniqueIndex;not null"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	Email     string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	AvatarUrl *string   `json:"avatar_url" gorm:"column:avatar_url;size:255"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

func (PsUsers) TableName() string {
	return "ps_users"
}

// PsUsedQuota represents the ps_used_quota table
type PsUsedQuota struct {
	UserId      uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey;constraint:OnDelete:CASCADE"`
	UsedQuota   int64     `json:"used_quota" gorm:"column:used_quota;default:0;not null"` // in bytes
	LastUpdated time.Time `json:"last_updated" gorm:"column:last_updated;default:CURRENT_TIMESTAMP"`

	// Relationships
	User PsUsers `gorm:"foreignKey:UserId;references:ID"`
}

func (PsUsedQuota) TableName() string {
	return "ps_used_quota"
}

// PsShares represents the ps_shares table
type PsShares struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserId        uuid.UUID  `json:"user_id" gorm:"column:user_id;type:uuid;not null;constraint:OnDelete:CASCADE"`
	CreatedAt     time.Time  `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt     *time.Time `json:"deleted_at" gorm:"column:deleted_at"`
	Title         string     `json:"title" gorm:"size:255;not null"`
	Description   *string    `json:"description" gorm:"type:text"`
	FileCount     int        `json:"file_count" gorm:"column:file_count;default:0"`
	Size          int64      `json:"size" gorm:"default:0"`
	DownloadCount int        `json:"download_count" gorm:"column:download_count;default:0"`
	ViewCount     int        `json:"view_count" gorm:"column:view_count;default:0"`
	IsPublic      bool       `json:"is_public" gorm:"column:is_public;default:true"`

	// Relationships
	User  PsUsers   `gorm:"foreignKey:UserId;references:ID"`
	Files []PsFiles `gorm:"foreignKey:ShareId;references:ID"`
}

func (PsShares) TableName() string {
	return "ps_shares"
}

// PsFiles represents the ps_files table
type PsFiles struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at"`
	S3Key     *string    `json:"s3_key" gorm:"column:s3_key;size:512"`
	ShareId   uuid.UUID  `json:"share_id" gorm:"column:share_id;type:uuid;not null;constraint:OnDelete:CASCADE"`
	FileName  string     `json:"file_name" gorm:"column:file_name;size:255;not null"`
	Mimetype  string     `json:"mimetype" gorm:"size:100;not null"`
	Hash      string     `json:"hash" gorm:"size:255;not null"`
	Size      int64      `json:"size" gorm:"not null"`

	// Relationships
	Share PsShares `gorm:"foreignKey:ShareId;references:ID"`
}

func (PsFiles) TableName() string {
	return "ps_files"
}

// PsUploadSignatures represents the ps_upload_signatures table
type PsUploadSignatures struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ShareId           uuid.UUID  `json:"share_id" gorm:"column:share_id;type:uuid;not null;constraint:OnDelete:CASCADE"`
	Signature         string     `json:"signature" gorm:"size:512;uniqueIndex;not null"`
	Expiry            time.Time  `json:"expiry" gorm:"not null"`
	IsUsed            bool       `json:"is_used" gorm:"column:is_used;default:false;not null"`
	ExpectedFileCount int        `json:"expected_file_count" gorm:"column:expected_file_count;not null"`
	ExpectedFileSize  int64      `json:"expected_file_size" gorm:"column:expected_file_size;not null"` // in MB
	UsedAt            *time.Time `json:"used_at" gorm:"column:used_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`

	// Relationships
	Share PsShares `gorm:"foreignKey:ShareId;references:ID"`
}

func (PsUploadSignatures) TableName() string {
	return "ps_upload_signatures"
}

// PsDownloadAnalytics represents the ps_download_analytics table
type PsDownloadAnalytics struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ShareId   uuid.UUID  `json:"share_id" gorm:"column:share_id;type:uuid;not null;constraint:OnDelete:CASCADE"`
	FileId    *uuid.UUID `json:"file_id" gorm:"column:file_id;type:uuid;constraint:OnDelete:SET NULL"`
	Timestamp time.Time  `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`
	IpAddress *string    `json:"ip_address" gorm:"column:ip_address;size:45"`
	UserAgent *string    `json:"user_agent" gorm:"column:user_agent;size:512"`
	Country   *string    `json:"country" gorm:"size:2"`
	City      *string    `json:"city" gorm:"size:100"`

	// Relationships
	Share PsShares `gorm:"foreignKey:ShareId;references:ID"`
	File  *PsFiles `gorm:"foreignKey:FileId;references:ID"`
}

func (PsDownloadAnalytics) TableName() string {
	return "ps_download_analytics"
}

// PsVisitAnalytics represents the ps_visit_analytics table
type PsVisitAnalytics struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ShareId   uuid.UUID `json:"share_id" gorm:"column:share_id;type:uuid;not null;constraint:OnDelete:CASCADE"`
	Timestamp time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`
	IpAddress *string   `json:"ip_address" gorm:"column:ip_address;size:45"`
	UserAgent *string   `json:"user_agent" gorm:"column:user_agent;size:512"`
	Referrer  *string   `json:"referrer" gorm:"size:512"`
	Country   *string   `json:"country" gorm:"size:2"`
	City      *string   `json:"city" gorm:"size:100"`

	// Relationships
	Share PsShares `gorm:"foreignKey:ShareId;references:ID"`
}

func (PsVisitAnalytics) TableName() string {
	return "ps_visit_analytics"
}
