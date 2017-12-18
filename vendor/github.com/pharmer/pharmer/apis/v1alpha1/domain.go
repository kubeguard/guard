package v1alpha1

type NameGenerator interface {
	ExtraNames(cluster string) []string
}

type NullNameGenerator struct {
}

var _ NameGenerator = &NullNameGenerator{}

func (NullNameGenerator) ExtraNames(cluster string) []string { return nil }
