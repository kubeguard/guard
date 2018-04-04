//go:generate go-enum -f=authchoice.go --lower
package ldap

// AuthChoice x ENUM(
// Simple,
// Kerberos
// )
type AuthChoice int32

// https://github.com/spf13/pflag/blob/1ce0cc6db4029d97571db82f85092fccedb572ce/flag.go#L187:6
func (e *AuthChoice) Set(name string) error {
	v, err := ParseAuthChoice(name)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

func (AuthChoice) Type() string {
	return "AuthChoice"
}
