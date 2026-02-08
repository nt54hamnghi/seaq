// TESTME:

package config

import (
	"errors"
	"fmt"

	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/spf13/viper"
)

var ErrUnexpectedType = errors.New("unexpected type of connections")

func AddConnection(conn llm.Connection) error {
	cs, err := llm.GetConnectionSet()
	if err != nil {
		return err
	}

	if cs.Has(conn.Provider) {
		return fmt.Errorf("connection %q already exists", conn.Provider)
	}

	viper.Set("connections", append(cs.AsSlice(), conn))
	return viper.WriteConfig()
}

func RemoveConnection(providers ...string) error {
	cs, err := llm.GetConnectionSet()
	if err != nil {
		return err
	}

	for _, p := range providers {
		if !cs.Has(p) {
			return fmt.Errorf("connection %q not found", p)
		}
		cs.Delete(p)
		fmt.Println(p)
	}

	viper.Set("connections", cs.AsSlice())
	return viper.WriteConfig()
}

func ListConnections() ([]llm.Connection, error) {
	cs, err := llm.GetConnectionSet()
	if err != nil {
		return nil, err
	}
	return cs.AsSlice(), nil
}
