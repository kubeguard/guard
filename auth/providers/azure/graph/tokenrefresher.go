package graph

// TokenRefresher is an interface to obtain token for MS Graph api
type TokenRefresher interface {
	Name() string
	Refresh(token string) (AuthResponse, error)
}
