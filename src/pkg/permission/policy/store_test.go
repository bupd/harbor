// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/permission/policy"
)

// The point of the change: what the five roles grant from the database is
// exactly what they granted from the map compiled into the binary. Every role
// is checked against every permission any of them can hold, so a grant that
// appeared or went missing shows up here rather than in production.
func TestSeededPolicyGrantsExactlyWhatTheMapGranted(t *testing.T) {
	store, err := policy.NewInMemory(policy.Shipped())
	require.NoError(t, err)

	want := map[string]bool{}
	for _, g := range policy.Shipped() {
		want[policy.Subject(g.RoleID)+"|"+g.Resource+"|"+g.Action] = true
	}

	vocabulary := policy.ProjectPolicies()
	assert.NotEmpty(t, vocabulary)

	for roleID := int64(1); roleID <= 5; roleID++ {
		subject := policy.Subject(roleID)
		for _, p := range vocabulary {
			resource := p.Resource.String()
			action := p.Action.String()

			granted, err := store.Enforce(subject, policy.Object(resource), action)
			require.NoError(t, err)
			assert.Equal(t, want[subject+"|"+resource+"|"+action], granted,
				"role %d, %s:%s", roleID, resource, action)
		}
	}
}

// keyMatch2 expands /* to everything below it and :pid to one path segment.
// Writing the stored object the first way would hand every role read access to
// every subresource of a project it can see.
func TestAGrantOnTheProjectDoesNotReachItsSubresources(t *testing.T) {
	store, err := policy.NewInMemory(policy.Shipped())
	require.NoError(t, err)

	granted, err := store.Enforce(policy.Subject(3), policy.Object(""), "read")
	require.NoError(t, err)
	assert.True(t, granted, "guest reads the project itself")

	granted, err = store.Enforce(policy.Subject(3), policy.Object("metadata"), "read")
	require.NoError(t, err)
	assert.False(t, granted, "which is not permission to read its metadata")
}

func TestAnUnknownRoleGrantsNothing(t *testing.T) {
	store, err := policy.NewInMemory(policy.Shipped())
	require.NoError(t, err)

	granted, err := store.Enforce(policy.Subject(9999), policy.Object("repository"), "pull")
	require.NoError(t, err)
	assert.False(t, granted)
}

// The store the process holds before the database is up has to deny, not panic
// and not allow.
func TestAnEmptyStoreDeniesEverything(t *testing.T) {
	store, err := policy.Empty()
	require.NoError(t, err)

	granted, err := store.Enforce(policy.Subject(1), policy.Object("repository"), "pull")
	require.NoError(t, err)
	assert.False(t, granted)
}

// The counts are here so that adding or dropping a permission in the map shows
// up as a deliberate change rather than a silent one. projectAdmin gained
// metadata:list in #23804.
func TestTheShippedRolesGrantWhatTheyAlwaysDid(t *testing.T) {
	counts := map[int64]int{}
	for _, g := range policy.Shipped() {
		counts[g.RoleID]++
	}
	assert.Equal(t, 76, counts[1], "projectAdmin")
	assert.Equal(t, 36, counts[2], "developer")
	assert.Equal(t, 20, counts[3], "guest")
	assert.Equal(t, 53, counts[4], "maintainer")
	assert.Equal(t, 14, counts[5], "limitedGuest")
}
