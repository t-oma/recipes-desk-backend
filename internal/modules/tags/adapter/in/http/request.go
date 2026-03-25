package httphandler

// SearchTagsRequest represents the request for listing tags.
type SearchTagsRequest struct {
	Query string `form:"q"`                // Search query for autocomplete
	Page  int    `form:"page,default=1"`   // Pagination page
	Limit int    `form:"limit,default=20"` // Pagination limit
}

// CreateTagRequest represents the create tag request.
type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}
