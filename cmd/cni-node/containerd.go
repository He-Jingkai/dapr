package main

import (
	"context"
	"fmt"
	ctr "github.com/containerd/go-runc"
)

func GetContainerByID(id string) (*ctr.Container, error) {
	ctrRunc := &ctr.Runc{}
	list, err := ctrRunc.List(context.Background())
	if err != nil {
		return nil, err
	}
	fmt.Printf("get ctrs %v\n", list)
	for _, container := range list {
		if container.ID == id {
			return container, nil
		}
	}
	return nil, fmt.Errorf("no containerd named %s", id)
}
