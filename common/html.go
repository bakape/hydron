package common

// Carries meta data an HTML page
type Page struct {
	SearchParams     string // Search tags and parameters
	Viewing          *Image // Image currently being viewed, if any
	Page, TotalPages int
}
