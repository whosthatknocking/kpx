package vault

import (
	"fmt"
	"sort"

	"github.com/tobischo/gokeepasslib/v3"
	"github.com/whosthatknocking/kpx/internal/cli"
)

// ListGroups returns all nested group paths, excluding the synthetic root group.
func (v *Vault) ListGroups() []string {
	root := v.rootGroup()
	paths := make([]string, 0)
	var walk func(parentPath string, group *gokeepasslib.Group)
	walk = func(parentPath string, group *gokeepasslib.Group) {
		for i := range group.Groups {
			child := &group.Groups[i]
			childPath := joinGroupPath(parentPath, child.Name)
			paths = append(paths, childPath)
			walk(childPath, child)
		}
	}
	walk("/", root)
	sort.Strings(paths)
	return paths
}

// AddGroup creates each missing path segment under the root group.
func (v *Vault) AddGroup(groupPath string) error {
	segments := splitGroupPath(groupPath)
	if len(segments) == 0 {
		return cli.NewExitError(cli.ExitGeneric, "group path must not be /")
	}

	current := v.rootGroup()
	for _, segment := range segments {
		child, err := findUniqueChildGroup(current, segment)
		if err != nil {
			return err
		}
		if child == nil {
			next := gokeepasslib.NewGroup()
			next.Name = segment
			current.Groups = append(current.Groups, next)
			child = &current.Groups[len(current.Groups)-1]
		}
		current = child
	}
	return nil
}

func (v *Vault) groupByPath(groupPath string) (*gokeepasslib.Group, error) {
	segments := splitGroupPath(groupPath)
	current := v.rootGroup()
	for _, segment := range segments {
		next, err := findUniqueChildGroup(current, segment)
		if err != nil {
			return nil, err
		}
		if next == nil {
			return nil, cli.NewExitError(cli.ExitNotFound, fmt.Sprintf("group not found: %s", normalizeGroupPath(groupPath)))
		}
		current = next
	}
	return current, nil
}

func findUniqueChildGroup(parent *gokeepasslib.Group, name string) (*gokeepasslib.Group, error) {
	var match *gokeepasslib.Group
	for i := range parent.Groups {
		child := &parent.Groups[i]
		if child.Name != name {
			continue
		}
		if match != nil {
			return nil, cli.NewExitError(cli.ExitAmbiguous, fmt.Sprintf("multiple groups matched %q", name))
		}
		match = child
	}
	return match, nil
}
