package project

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/bolt"
	"github.com/ansible-semaphore/semaphore/util"
	"testing"
)

type testItem struct {
	Name string
}

func TestBackupProject(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	store := bolt.CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Name: "Test 123",
	})
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	if err != nil {
		t.Fatal(err)
	}

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        1,
	})
	if err != nil {
		t.Fatal(err)
	}

	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		RepositoryID:  repo.ID,
		InventoryID:   &inv.ID,
		EnvironmentID: &env.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	backup, err := GetBackup(proj.ID, store)
	if err != nil {
		t.Fatal(err)
	}

	if backup.Meta.ID != proj.ID {
		t.Fatal("backup meta ID wrong")
	}

	str, err := backup.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if str != `{"environments":[{"env":null,"json":"{\"author\": \"Denis\", \"comment\": \"Hello, World!\"}","name":"test","password":null}],"inventories":[{"become_key":null,"inventory":"","name":"","ssh_key":null,"type":""}],"keys":[{"name":"","type":"none"}],"meta":{"alert":false,"alert_chat":null,"max_parallel_tasks":0,"name":"Test 123","type":""},"repositories":[{"git_branch":"master","git_url":"git@example.com:test/test","name":"Test","ssh_key":""}],"templates":[{"allow_override_args_in_task":false,"app":"","arguments":null,"autorun":false,"build_template":null,"cron":null,"description":null,"environment":"test","inventory":"","name":"Test","playbook":"test.yml","repository":"Test","start_version":null,"suppress_success_alerts":false,"survey_vars":"null","type":"","vaults":null,"view":null}],"views":null}` {
		t.Fatal("Invalid backup content")
	}

	restoredBackup := &BackupFormat{}
	err = restoredBackup.Unmarshal(str)
	if err != nil {
		t.Fatal(err)
	}

	if restoredBackup.Meta.Name != proj.Name {
		t.Fatal("backup meta ID wrong")
	}

}

func isUnique(items []testItem) bool {
	for i, item := range items {
		for k, other := range items {
			if i == k {
				continue
			}

			if item.Name == other.Name {
				return false
			}
		}
	}

	return true
}

func TestMakeUniqueNames(t *testing.T) {
	items := []testItem{
		{Name: "Project"},
		{Name: "Solution"},
		{Name: "Project"},
		{Name: "Project"},
		{Name: "Project"},
		{Name: "Project"},
	}

	makeUniqueNames(items, func(item *testItem) string {
		return item.Name
	}, func(item *testItem, name string) {
		item.Name = name
	})

	if !isUnique(items) {
		t.Fatal("Not unique names")
	}
}
