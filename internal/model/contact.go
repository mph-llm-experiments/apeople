package model

import (
	"time"
)

// RelationshipType defines the type of relationship and default contact frequency
type RelationshipType string

const (
	RelationshipClose      RelationshipType = "close"      // 30 days
	RelationshipFamily     RelationshipType = "family"     // 30 days
	RelationshipNetwork    RelationshipType = "network"    // 90 days
	RelationshipWork       RelationshipType = "work"       // 60 days
	RelationshipSocial     RelationshipType = "social"     // No default
	RelationshipProviders  RelationshipType = "providers"  // No default
	RelationshipRecruiters RelationshipType = "recruiters" // No default
)

// ContactStyle defines how contact reminders work
type ContactStyle string

const (
	StylePeriodic  ContactStyle = "periodic"  // Regular check-ins
	StyleAmbient   ContactStyle = "ambient"   // Passive monitoring
	StyleTriggered ContactStyle = "triggered" // Event-based
)

// ContactState represents the current state of a contact
type ContactState string

const (
	StateActive   ContactState = "active"
	StateFollowup ContactState = "followup"
	StatePing     ContactState = "ping"
	StateArchived ContactState = "archived"
	StateOk       ContactState = "ok"
)

// InteractionType represents types of interactions
type InteractionType string

const (
	InteractionEmail   InteractionType = "email"
	InteractionCall    InteractionType = "call"
	InteractionText    InteractionType = "text"
	InteractionMeeting InteractionType = "meeting"
	InteractionSocial  InteractionType = "social"
	InteractionBump    InteractionType = "bump"
	InteractionNote    InteractionType = "note"
)

// Contact represents a contact record
type Contact struct {
	// Core fields from frontmatter
	Title            string           `yaml:"title" json:"title"`
	Date             time.Time        `yaml:"date" json:"date"`
	Tags             []string         `yaml:"tags" json:"tags"`
	Identifier       string           `yaml:"identifier" json:"identifier"`
	IndexID          int              `yaml:"index_id,omitempty" json:"index_id"`
	Email            string           `yaml:"email,omitempty" json:"email,omitempty"`
	Phone            string           `yaml:"phone,omitempty" json:"phone,omitempty"`
	RelationshipType RelationshipType `yaml:"relationship_type" json:"relationship_type"`
	State            string           `yaml:"state,omitempty" json:"state,omitempty"`
	Label            string           `yaml:"label,omitempty" json:"label,omitempty"`
	ContactStyle     ContactStyle     `yaml:"contact_style,omitempty" json:"contact_style,omitempty"`
	LastContacted    *time.Time       `yaml:"last_contacted,omitempty" json:"last_contacted,omitempty"`
	LastBumpDate     *time.Time       `yaml:"last_bump_date,omitempty" json:"last_bump_date,omitempty"`
	BumpCount        int              `yaml:"bump_count,omitempty" json:"bump_count,omitempty"`
	UpdatedAt        time.Time        `yaml:"updated_at" json:"updated_at"`

	// Optional fields
	Company              string   `yaml:"company,omitempty" json:"company,omitempty"`
	Role                 string   `yaml:"role,omitempty" json:"role,omitempty"`
	Location             string   `yaml:"location,omitempty" json:"location,omitempty"`
	Birthday             string   `yaml:"birthday,omitempty" json:"birthday,omitempty"`
	LinkedIn             string   `yaml:"linkedin,omitempty" json:"linkedin,omitempty"`
	Twitter              string   `yaml:"twitter,omitempty" json:"twitter,omitempty"`
	Website              string   `yaml:"website,omitempty" json:"website,omitempty"`
	Notes                string   `yaml:"notes,omitempty" json:"notes,omitempty"`
	CustomFrequencyDays  int      `yaml:"custom_frequency_days,omitempty" json:"custom_frequency_days,omitempty"`
	LastInteractionType  string   `yaml:"last_interaction_type,omitempty" json:"last_interaction_type,omitempty"`
	RelatedContactLabels []string `yaml:"related_contact_labels,omitempty" json:"related_contact_labels,omitempty"`

	// Cross-app relationship fields (asystem connective tissue)
	RelatedPeople []string `yaml:"related_people,omitempty" json:"related_people"`
	RelatedTasks  []string `yaml:"related_tasks,omitempty" json:"related_tasks"`
	RelatedIdeas  []string `yaml:"related_ideas,omitempty" json:"related_ideas"`

	// Runtime/computed fields (not in YAML)
	FilePath         string `yaml:"-" json:"file_path,omitempty"`
	Content          string `yaml:"-" json:"-"`
	DaysSince        int    `yaml:"-" json:"days_since_contact"`
	OverdueStatus    string `yaml:"-" json:"overdue_status,omitempty"`
}

// Interaction represents a single interaction with a contact
type Interaction struct {
	Date    time.Time       `yaml:"date"`
	Type    InteractionType `yaml:"type"`
	Summary string          `yaml:"summary,omitempty"`
}

// EnsureRelationSlices initializes nil relation slices to empty slices
// so JSON output shows [] instead of null.
func (c *Contact) EnsureRelationSlices() {
	if c.RelatedPeople == nil {
		c.RelatedPeople = []string{}
	}
	if c.RelatedTasks == nil {
		c.RelatedTasks = []string{}
	}
	if c.RelatedIdeas == nil {
		c.RelatedIdeas = []string{}
	}
}

// GetFrequencyDays returns the contact frequency in days
func (c *Contact) GetFrequencyDays() int {
	if c.CustomFrequencyDays > 0 {
		return c.CustomFrequencyDays
	}

	switch c.RelationshipType {
	case RelationshipClose, RelationshipFamily:
		return 30
	case RelationshipWork:
		return 60
	case RelationshipNetwork:
		return 90
	default:
		return 0 // No default for social
	}
}

// DaysSinceContact returns days since last contact (not bump)
func (c *Contact) DaysSinceContact() int {
	if c.LastContacted == nil {
		return -1 // Never contacted
	}
	duration := time.Since(*c.LastContacted)
	days := int(duration.Hours() / 24)
	// Handle future dates (negative days)
	if duration < 0 {
		return days // Will be negative
	}
	return days
}

// IsOverdue returns true if contact is overdue based on frequency
func (c *Contact) IsOverdue() bool {
	// Only check overdue for periodic style
	if c.ContactStyle != StylePeriodic && c.ContactStyle != "" {
		return false
	}
	
	freq := c.GetFrequencyDays()
	if freq == 0 {
		return false // No frequency set
	}
	
	days := c.DaysSinceContact()
	if days == -1 {
		return true // Never contacted
	}
	
	return days > freq
}

// NeedsAttention returns true if contact needs attention soon
func (c *Contact) NeedsAttention() bool {
	// Only check for periodic style
	if c.ContactStyle != StylePeriodic && c.ContactStyle != "" {
		return false
	}
	
	freq := c.GetFrequencyDays()
	if freq == 0 {
		return false
	}
	
	days := c.DaysSinceContact()
	if days == -1 {
		return true
	}
	
	// Needs attention if within 7 days of being overdue
	return days > (freq - 7) && days <= freq
}

// IsWithinThreshold returns true if contact has been contacted within their expected frequency
func (c *Contact) IsWithinThreshold() bool {
	// Only check for periodic style
	if c.ContactStyle != StylePeriodic && c.ContactStyle != "" {
		return false
	}
	
	freq := c.GetFrequencyDays()
	if freq == 0 {
		return false
	}
	
	days := c.DaysSinceContact()
	if days == -1 {
		return false // Never contacted
	}
	
	// Within threshold if contacted recently enough (less than half the frequency)
	// This gives a nice visual indicator for "good" contact rhythm
	return days >= 0 && days <= (freq / 2)
}