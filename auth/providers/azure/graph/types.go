package graph

// AuthResponse represents a response from the MS Graph auth API
type AuthResponse struct {
	TokenType string `json:"token_type"`
	Expires   int    `json:"expires_in"`
	Token     string `json:"access_token"`
}

// NOTE: These below are partial implementations of the API objects containing
// only the necessary fields to perform the functions of this package

// ObjectList represents a list of directory object IDs returned from the MS Graph API
type ObjectList struct {
	Value []string `json:"value"`
}

// ObjectQuery represents a query object to the directoryObjects endpoint
type ObjectQuery struct {
	IDs   []string `json:"ids"`
	Types []string `json:"types"`
}

// GroupList represents a list of groups returned from the MS Graph API
type GroupList struct {
	Value []Group `json:"value"`
}

// Group represents the Group object from the MSGraphAPI
type Group struct {
	Name string `json:"displayName"`
	ID   string `json:"id"`
}
