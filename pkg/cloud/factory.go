package cloud

import (
	"errors"
	"fmt"

	"github.com/MovieStoreGuy/skirmish/pkg/minions"
)

// Factory defines a map of function pointers
type Factory map[string]func() minions.Minion

func (fact Factory) CreateMinion(name string) (minions.Minion, error) {
	if fact == nil {
		return nil, errors.New("factory is not initialised")
	}
	op, exist := (fact)[name]
	if !exist {
		return nil, fmt.Errorf("%s minion not registered", name)
	}
	return op(), nil
}

func (fact Factory) RegisterMinion(name string, op func() minions.Minion) error {
	if fact == nil {
		return errors.New("factory is not initialised")
	}
	fact[name] = op
	return nil
}
