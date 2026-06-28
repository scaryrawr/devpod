package platform

import (
	"fmt"

	"github.com/spf13/pflag"
)

type OwnerFilter string

const (
	SelfOwnerFilter OwnerFilter = "self"
	AllOwnerFilter  OwnerFilter = "all"
)

var _ pflag.Value = (*OwnerFilter)(nil)

func (s *OwnerFilter) Set(v string) error {
	switch v {
	case "":
		{
			*s = SelfOwnerFilter
			return nil
		}
	case string(SelfOwnerFilter),
		string(AllOwnerFilter):
		{
			*s = OwnerFilter(v)
			return nil
		}
	default:
		return fmt.Errorf("OwnerFilter %s not supported", v)
	}
}

func (s *OwnerFilter) Type() string {
	return "ownerFilter"
}

func (s *OwnerFilter) String() string {
	return string(*s)
}
