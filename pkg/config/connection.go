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
	conns, err := llm.GetConnections()
	if err != nil {
		return err
	}

	if _, ok := conns[conn.Provider]; ok {
		return fmt.Errorf("connection %q already exists", conn.Provider)
	}

	viper.Set("connections", append(conns.AsSlice(), conn))
	return viper.WriteConfig()
}

func RemoveConnection(providers ...string) error {
	conns, err := llm.GetConnections()
	if err != nil {
		return err
	}

	for _, p := range providers {
		if _, ok := conns[p]; !ok {
			return fmt.Errorf("connection %q not found", p)
		}
		delete(conns, p)
		fmt.Println(p)
	}

	viper.Set("connections", conns.AsSlice())
	return viper.WriteConfig()
}

func ListConnections() ([]llm.Connection, error) {
	conns, err := llm.GetConnections()
	if err != nil {
		return nil, err
	}
	return conns.AsSlice(), nil
}
