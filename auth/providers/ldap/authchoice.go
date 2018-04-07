//go:generate go-enum -f=authchoice.go --lower --flag
package ldap

// AuthChoice x ENUM(
// Simple,
// Kerberos
// )
type AuthChoice int32
