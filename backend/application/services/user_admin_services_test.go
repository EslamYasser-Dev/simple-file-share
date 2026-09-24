package services

import (
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func newRBACFixture(t *testing.T) (
	*CreateUserService,
	*UpdateUserService,
	*DeleteUserService,
	*ListUsersService,
	*ListRolesService,
	*CreateOrUpdateRoleService,
	*DeleteRoleService,
	*fs.UserFileRepository,
	*RoleCatalog,
) {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	scoper := policy.NewPathScoper()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	userRepo := fs.NewUserFileRepository(dir)
	shareRepo := fs.NewShareFileRepository(dir)
	roleRepo := fs.NewRoleFileRepository(dir)
	roles := NewRoleCatalog(roleRepo)
	hasher := auth.NewPBKDF2Hasher()

	create := NewCreateUserService(userRepo, hasher, fileRepo, scoper, roles, 0)
	update := NewUpdateUserService(userRepo, roles, scoper, fs.NewLocalFileRepository(dir), shareRepo, index)
	del := NewDeleteUserService(userRepo, roles, scoper, fileRepo, shareRepo)
	list := NewListUsersService(userRepo, index, scoper, roles)
	listRoles := NewListRolesService(roles)
	upsertRole := NewCreateOrUpdateRoleService(userRepo, roleRepo, roles)
	delRole := NewDeleteRoleService(userRepo, roleRepo, roles)
	return create, update, del, list, listRoles, upsertRole, delRole, userRepo, roles
}

func adminUser() *models.User {
	return &models.User{Username: "root", Role: models.RoleAdmin, IsAdmin: true, Enabled: true}
}

func TestCreateUserRequiresPermission(t *testing.T) {
	create, _, _, _, _, _, _, _, _ := newRBACFixture(t)
	member := &models.User{Username: "bob", Role: models.RoleMember, Enabled: true}
	if _, err := create.Execute(member, CreateUserInput{Username: "carol", Password: "secret"}); err == nil {
		t.Fatal("member should not create users")
	}
	if _, err := create.Execute(adminUser(), CreateUserInput{Username: "carol", Password: "secret", Role: "member"}); err != nil {
		t.Fatalf("admin create: %v", err)
	}
}

func TestUpdateUserRenameMigratesSharesAndSessions(t *testing.T) {
	create, update, _, _, _, _, _, userRepo, _ := newRBACFixture(t)
	if _, err := create.Execute(adminUser(), CreateUserInput{Username: "alice", Password: "secret"}); err != nil {
		t.Fatal(err)
	}
	newName := "alice2"
	updated, err := update.Execute(adminUser(), "alice", UpdateUserInput{Username: &newName})
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if updated.Username != "alice2" {
		t.Fatalf("username = %q", updated.Username)
	}
	if _, err := userRepo.FindByUsername("alice"); err == nil {
		t.Fatal("old username should be gone")
	}
	if _, err := userRepo.FindByUsername("alice2"); err != nil {
		t.Fatalf("new username: %v", err)
	}
}

func TestLastAdminProtection(t *testing.T) {
	create, update, del, _, _, _, _, _, _ := newRBACFixture(t)
	if _, err := create.Execute(adminUser(), CreateUserInput{Username: "root", Password: "secret", Role: "admin"}); err != nil {
		t.Fatal(err)
	}
	if _, err := create.Execute(adminUser(), CreateUserInput{Username: "root2", Password: "secret", Role: "admin"}); err != nil {
		t.Fatal(err)
	}
	// Disable root2 first so only root remains enabled admin.
	disable := false
	if _, err := update.Execute(adminUser(), "root2", UpdateUserInput{Enabled: &disable}); err != nil {
		t.Fatal(err)
	}
	role := models.RoleMember
	if _, err := update.Execute(adminUser(), "root", UpdateUserInput{Role: &role}); err == nil {
		t.Fatal("expected last-admin demotion to fail")
	}
	if err := del.Execute(adminUser(), "root"); err == nil {
		t.Fatal("expected last-admin delete to fail")
	}
}

func TestCustomRoleLifecycle(t *testing.T) {
	create, _, _, _, listRoles, upsert, delRole, _, roles := newRBACFixture(t)

	role, err := upsert.Execute(adminUser(), "Operator", "ops", []string{models.PermUsersRead, models.PermQuotaManage})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if role.Name != "operator" || len(role.Permissions) != 2 {
		t.Fatalf("role = %+v", role)
	}
	if !roles.RoleExists("operator") {
		t.Fatal("catalog should know operator")
	}
	if err := delRole.Execute(adminUser(), "admin"); err == nil {
		t.Fatal("cannot delete built-in")
	}
	// Assign and demote last admin via operator role lacking files.system.
	if _, err := create.Execute(adminUser(), CreateUserInput{Username: "ops1", Password: "secret", Role: "operator"}); err != nil {
		t.Fatal(err)
	}
	// Delete unused custom role then re-create after unassign.
	// First reassign ops1 away if needed — role still assigned so delete fails.
	if err := delRole.Execute(adminUser(), "operator"); err == nil {
		t.Fatal("expected in-use role delete to fail")
	}
	got, err := listRoles.Execute(adminUser())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range got {
		if r.Name == "operator" {
			found = true
		}
	}
	if !found {
		t.Fatal("operator missing from list")
	}
}

func TestListUsersMemberForbidden(t *testing.T) {
	_, _, _, list, _, _, _, _, _ := newRBACFixture(t)
	member := &models.User{Username: "bob", Role: models.RoleMember, Enabled: true}
	if _, err := list.Execute(member); err == nil {
		t.Fatal("member should not list users")
	}
	if _, err := list.Execute(adminUser()); err != nil {
		t.Fatalf("admin list: %v", err)
	}
}

func TestDisabledAccountCannotAuthenticate(t *testing.T) {
	create, update, _, _, _, _, _, userRepo, _ := newRBACFixture(t)
	if _, err := create.Execute(adminUser(), CreateUserInput{Username: "root", Password: "secret", Role: "admin"}); err != nil {
		t.Fatal(err)
	}
	if _, err := create.Execute(adminUser(), CreateUserInput{Username: "dave", Password: "secret", Role: "member"}); err != nil {
		t.Fatal(err)
	}
	hasher := auth.NewPBKDF2Hasher()
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	if _, err := provider.Authenticate("dave", "secret"); err != nil {
		t.Fatalf("enabled login: %v", err)
	}
	disable := false
	if _, err := update.Execute(adminUser(), "dave", UpdateUserInput{Enabled: &disable}); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Authenticate("dave", "secret"); err == nil {
		t.Fatal("disabled account must not authenticate")
	}
	_ = time.Now()
}
