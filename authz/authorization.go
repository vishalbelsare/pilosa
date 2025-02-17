// Copyright 2022 Molecula Corp. (DBA FeatureBase).
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/featurebasedb/featurebase/v3/authn"

	"gopkg.in/yaml.v2"
)

type GroupPermissions struct {
	Permissions map[string]map[string]Permission `yaml:"user-groups"`
	Admin       string                           `yaml:"admin"`
}

type Permission string

const (
	None  Permission = ""
	Read  Permission = "read"
	Write Permission = "write"
	Admin Permission = "admin"
)

// Satisfies returns whether `p` satisfies the permissions required by `b`
func (p Permission) Satisfies(b Permission) bool {
	switch p {
	case "":
		return b == ""
	case "read":
		return b == "" || b == "read"
	case "write":
		return b == "" || b == "read" || b == "write"
	case "admin":
		return b == "" || b == "read" || b == "write" || b == "admin"
	}
	return false
}

func (p *GroupPermissions) ReadPermissionsFile(permsFile io.Reader) (err error) {
	permsData, err := ioutil.ReadAll(permsFile)

	if err != nil {
		return fmt.Errorf("reading permissions failed with error: %s", err)
	}

	err = yaml.UnmarshalStrict(permsData, &p)
	if err != nil {
		return fmt.Errorf("unmarshalling permissions failed with error: %s", err)
	}

	return
}

func (p *GroupPermissions) GetPermissions(user *authn.UserInfo, index string) (permission Permission, errors error) {
	groups := user.Groups
	if admin := p.IsAdmin(groups); admin {
		return Admin, nil
	}

	allPermissions := map[Permission]bool{
		Write: false,
		Read:  false,
	}

	if len(groups) == 0 {
		return None, fmt.Errorf("user is not part of any groups in identity provider")
	}

	var groupsDenied []string
	for _, group := range groups {
		if _, ok := p.Permissions[group.GroupID]; ok {
			if perm, ok := p.Permissions[group.GroupID][index]; ok {
				allPermissions[perm] = true
			} else {
				return None, fmt.Errorf("user %s does not have permission to index %s", user.UserID, index)
			}
		} else {
			groupsDenied = append(groupsDenied, group.GroupID)
		}
	}

	if len(groupsDenied) == len(groups) {
		return None, fmt.Errorf("group(s) %s does not have permission to FeatureBase", groupsDenied)
	}

	if allPermissions[Write] {
		return Write, nil
	} else if allPermissions[Read] {
		return Read, nil
	} else {
		return None, fmt.Errorf("no permissions found")
	}
}

func (p *GroupPermissions) IsAdmin(groups []authn.Group) bool {
	for _, group := range groups {
		if p.Admin == group.GroupID {
			return true
		}
	}
	return false
}

func (p *GroupPermissions) GetAuthorizedIndexList(groups []authn.Group, desiredPermission Permission) (indexList []string) {
	// if user is admin, find all indexes in permissions file and return them
	if p.IsAdmin(groups) {
		for groupId := range p.Permissions {
			for index := range p.Permissions[groupId] {
				indexList = append(indexList, index)
			}
		}
		return indexList
	}

	for _, group := range groups {
		if _, ok := p.Permissions[group.GroupID]; ok {
			for index, permission := range p.Permissions[group.GroupID] {
				if permission.Satisfies(desiredPermission) {
					indexList = append(indexList, index)
				}
			}
		}
	}
	return indexList
}
