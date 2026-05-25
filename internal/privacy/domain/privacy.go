package domain

import "time"

type RequestType string

const (
	RequestAccess      RequestType = "ACCESS"
	RequestCorrection  RequestType = "CORRECTION"
	RequestExport      RequestType = "EXPORT"
	RequestDeletion    RequestType = "DELETION"
	RequestPortability RequestType = "PORTABILITY"
)

type RequestStatus string

const (
	RequestOpen     RequestStatus = "OPEN"
	RequestInReview RequestStatus = "IN_REVIEW"
	RequestDone     RequestStatus = "DONE"
	RequestRejected RequestStatus = "REJECTED"
)

type PrivacyRequest struct {
	ID        uint
	UserID    uint
	Type      RequestType
	Status    RequestStatus
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
