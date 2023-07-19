/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/types/accesslist"
	"github.com/gravitational/teleport/lib/asciitable"
	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/service/servicecfg"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/trace"
)

// ACLCommand implements the `tctl acl` family of commands.
type ACLCommand struct {
	format string

	list        *kingpin.CmdClause
	show        *kingpin.CmdClause
	usersAdd    *kingpin.CmdClause
	usersRemove *kingpin.CmdClause
	usersList   *kingpin.CmdClause

	// Used for managing a particular access list.
	accessListName string

	// Used for managing membership to an access list.
	userName string
	expires  string
	reason   string
}

// Initialize allows ACLCommand to plug itself into the CLI parser
func (c *ACLCommand) Initialize(app *kingpin.Application, _ *servicecfg.Config) {
	acl := app.Command("acl", "Manage access lists.").Alias("access-lists")

	c.list = acl.Command("list", "List cluster access lists.").Alias("ls")
	c.list.Flag("format", "Output format, 'yaml', 'json', or 'text'").Default(teleport.YAML).EnumVar(&c.format, teleport.YAML, teleport.JSON, teleport.Text)

	c.show = acl.Command("show", "Show detailed information for an access list..")
	c.show.Arg("access-list-name", "The access list name.").Required().StringVar(&c.accessListName)
	c.show.Flag("format", "Output format, 'yaml', 'json', or 'text'").Default(teleport.YAML).EnumVar(&c.format, teleport.YAML, teleport.JSON, teleport.Text)

	users := acl.Command("users", "Manage user membership to access lists.")

	c.usersAdd = users.Command("add", "Add a user to an access list.")
	c.usersAdd.Arg("access-list-name", "The access list name.").Required().StringVar(&c.accessListName)
	c.usersAdd.Arg("user-name", "The user to add to the access list.").Required().StringVar(&c.userName)
	c.usersAdd.Arg("expires", "When the user's access expires (must be in RFC3339).").Required().StringVar(&c.expires)
	c.usersAdd.Arg("reason", "The reason the user has been added to the access list.").Required().StringVar(&c.reason)

	c.usersRemove = users.Command("rm", "Remove a user from an access list.")
	c.usersRemove.Arg("access-list-name", "The access list name.").Required().StringVar(&c.accessListName)
	c.usersRemove.Arg("user-name", "The user to add to the access list.").Required().StringVar(&c.userName)

	c.usersList = users.Command("ls", "List users that are members of an access list.")
	c.usersList.Arg("access-list-name", "The access list name.").Required().StringVar(&c.accessListName)
}

// TryRun takes the CLI command as an argument (like "acl ls") and executes it.
func (c *ACLCommand) TryRun(ctx context.Context, cmd string, client auth.ClientI) (match bool, err error) {
	switch cmd {
	case c.list.FullCommand():
		err = c.List(ctx, client)
	case c.show.FullCommand():
		err = c.Show(ctx, client)
	case c.usersAdd.FullCommand():
		err = c.UsersAdd(ctx, client)
	case c.usersRemove.FullCommand():
		err = c.UsersRemove(ctx, client)
	case c.usersList.FullCommand():
		err = c.UsersList(ctx, client)
	default:
		return false, nil
	}
	return true, trace.Wrap(err)
}

// List will list access lists visible to the user.
func (c *ACLCommand) List(ctx context.Context, client auth.ClientI) error {
	accessLists, err := client.AccessListClient().GetAccessLists(ctx)
	if err != nil {
		return trace.Wrap(err)
	}

	if len(accessLists) == 0 {
		fmt.Println("no access lists")
		return nil
	}

	return trace.Wrap(displayAccessLists(c.format, accessLists...))
}

// Show will display information about an access list visible to the user.
func (c *ACLCommand) Show(ctx context.Context, client auth.ClientI) error {
	accessList, err := client.AccessListClient().GetAccessList(ctx, c.accessListName)
	if err != nil {
		return trace.Wrap(err)
	}

	return trace.Wrap(displayAccessLists(c.format, accessList))
}

// UsersAdd will add a user to an access list.
func (c *ACLCommand) UsersAdd(ctx context.Context, client auth.ClientI) error {
	expires, err := time.Parse(time.RFC3339, c.expires)
	if err != nil {
		return trace.Wrap(err)
	}

	accessList, err := client.AccessListClient().GetAccessList(ctx, c.accessListName)
	if err != nil {
		return trace.Wrap(err)
	}

	accessList.Spec.Members = append(accessList.Spec.Members, accesslist.Member{
		Name:    c.userName,
		Reason:  c.reason,
		Expires: expires,

		// The following fields will be updated in the backend, so their values here don't matter.
		Joined:  time.Now(),
		AddedBy: "dummy",
	})

	_, err = client.AccessListClient().UpsertAccessList(ctx, accessList)
	if err != nil {
		return trace.Wrap(err)
	}

	fmt.Printf("successfully added user %s to access list %s", c.userName, c.accessListName)

	return nil
}

// UsersRemove will remove a user to an access list.
func (c *ACLCommand) UsersRemove(ctx context.Context, client auth.ClientI) error {
	accessList, err := client.AccessListClient().GetAccessList(ctx, c.accessListName)
	if err != nil {
		return trace.Wrap(err)
	}

	memberIndex := -1
	for i, member := range accessList.Spec.Members {
		if member.Name == c.userName {
			memberIndex = i
			break
		}
	}

	if memberIndex == -1 {
		return trace.NotFound("user %s is not a member of access list %s\n", c.userName, c.accessListName)
	}

	accessList.Spec.Members = append(accessList.Spec.Members[:memberIndex], accessList.Spec.Members[memberIndex+1:]...)

	_, err = client.AccessListClient().UpsertAccessList(ctx, accessList)
	if err != nil {
		return trace.Wrap(err)
	}

	fmt.Printf("successfully removed user %s from access list %s\n", c.userName, c.accessListName)

	return nil
}

// UsersList will list the users in an access list.
func (c *ACLCommand) UsersList(ctx context.Context, client auth.ClientI) error {
	accessList, err := client.AccessListClient().GetAccessList(ctx, c.accessListName)
	if err != nil {
		return trace.Wrap(err)
	}

	if len(accessList.Spec.Members) == 0 {
		fmt.Printf("No members found for access list %s.\nYou may not have access to see the members for this list.\n", c.accessListName)
		return nil
	}

	fmt.Printf("Members of %s:\n", c.accessListName)
	for _, member := range accessList.Spec.Members {
		fmt.Printf("- %s\n", member.Name)
	}

	return nil
}

func displayAccessLists(format string, accessLists ...*accesslist.AccessList) error {
	switch format {
	case teleport.YAML:
		return trace.Wrap(utils.WriteYAML(os.Stdout, accessLists))
	case teleport.JSON:
		return trace.Wrap(utils.WriteJSON(os.Stdout, accessLists))
	case teleport.Text:
		return trace.Wrap(displayAccessListsText(accessLists...))
	}

	// technically unreachable since kingpin validates the EnumVar
	return trace.BadParameter("invalid format %q", format)
}

func displayAccessListsText(accessLists ...*accesslist.AccessList) error {
	table := asciitable.MakeTable([]string{"ID", "Audit Frequency", "Granted Roles", "Granted Traits"})
	for _, accessList := range accessLists {
		grantedRoles := strings.Join(accessList.GetGrants().Roles, ",")
		traitStrings := make([]string, 0, len(accessList.GetGrants().Traits))
		for k, values := range accessList.GetGrants().Traits {
			traitStrings = append(traitStrings, fmt.Sprintf("%s:{%s}", k, strings.Join(values, ",")))
		}
		grantedTraits := strings.Join(traitStrings, ",")
		table.AddRow([]string{
			accessList.GetName(),
			accessList.GetAuditFrequency().String(),
			grantedRoles,
			grantedTraits,
		})
	}
	_, err := fmt.Println(table.AsBuffer().String())
	return trace.Wrap(err)
}
