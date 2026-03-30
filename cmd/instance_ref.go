package cmd

import (
	"fmt"
	"strings"

	"github.com/dims/lambdactl/api"
)

func resolveInstanceRef(c *api.Client, ref string) (*api.Instance, error) {
	instances, err := c.ListInstances()
	if err != nil {
		return nil, err
	}

	inst, err := findInstanceByRef(instances, ref)
	if err == nil {
		return inst, nil
	}

	if direct, directErr := c.GetInstance(ref); directErr == nil {
		return direct, nil
	}

	return nil, err
}

func findInstanceByRef(instances []api.Instance, ref string) (*api.Instance, error) {
	var idMatch *api.Instance
	var nameMatches []*api.Instance

	for i := range instances {
		inst := &instances[i]
		if inst.ID == ref {
			idMatch = inst
		}
		if inst.Name == ref && ref != "" {
			nameMatches = append(nameMatches, inst)
		}
	}

	if idMatch != nil {
		return idMatch, nil
	}
	if len(nameMatches) == 1 {
		return nameMatches[0], nil
	}
	if len(nameMatches) > 1 {
		ids := make([]string, 0, len(nameMatches))
		for _, inst := range nameMatches {
			ids = append(ids, inst.ID)
		}
		return nil, fmt.Errorf("multiple running instances named %q: %s", ref, strings.Join(ids, ", "))
	}

	return nil, fmt.Errorf("no running instance found with ID or name %q", ref)
}

func describeInstance(inst *api.Instance) string {
	if inst.Name != "" {
		return fmt.Sprintf("%q (%s)", inst.Name, inst.ID)
	}
	return inst.ID
}
