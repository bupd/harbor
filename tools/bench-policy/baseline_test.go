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

// This file is copied into a worktree of the base commit by run.sh, so that the
// same questions are asked of the unmodified code. It uses only what the base
// commit already exports.

package project

import (
	"context"
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	rbacevaluator "github.com/goharbor/harbor/src/pkg/permission/evaluator/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
)

var benchProject = &proModels.Project{
	ProjectID: 42,
	Name:      "bench",
	OwnerID:   1,
	Metadata:  map[string]string{"public": "false"},
}

// main holds resolved roles rather than ids, so the base side builds them.
func benchUser(roles ...int) types.RBACUser {
	var rs []*projectRBACRole
	for _, r := range roles {
		rs = append(rs, &projectRBACRole{projectID: benchProject.ProjectID, roleID: r})
	}
	return &rbacUser{project: benchProject, username: "alice", projectRoles: rs}
}

func BenchmarkHasPermission(b *testing.B) {
	resource := NewNamespace(benchProject.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := rbacevaluator.New(benchUser(common.RoleProjectAdmin))
		if !e.HasPermission(ctx, resource, rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}

func BenchmarkHasPermissionDenied(b *testing.B) {
	resource := NewNamespace(benchProject.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := rbacevaluator.New(benchUser(common.RoleGuest))
		if e.HasPermission(ctx, resource, rbac.ActionPush) {
			b.Fatal("permission granted")
		}
	}
}

func BenchmarkHasPermissionManyProjects(b *testing.B) {
	const projects = 10000
	resources := make([]types.Resource, projects)
	for i := range resources {
		resources[i] = NewNamespace(int64(i + 1)).Resource(rbac.ResourceRepository)
	}
	users := make([]types.RBACUser, projects)
	for i := range users {
		pid := int64(i + 1)
		users[i] = &rbacUser{
			project:      &proModels.Project{ProjectID: pid, Name: fmt.Sprintf("p%d", i), Metadata: map[string]string{"public": "false"}},
			username:     "alice",
			projectRoles: []*projectRBACRole{{projectID: pid, roleID: common.RoleProjectAdmin}},
		}
	}
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := i % projects
		e := rbacevaluator.New(users[n])
		if !e.HasPermission(ctx, resources[n], rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}

func BenchmarkHasPermissionPublicProject(b *testing.B) {
	public := &proModels.Project{ProjectID: 43, Name: "public", Metadata: map[string]string{"public": "true"}}
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := rbacevaluator.New(&rbacUser{project: public, username: "anonymous"})
		if !e.HasPermission(ctx, resource, rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}
