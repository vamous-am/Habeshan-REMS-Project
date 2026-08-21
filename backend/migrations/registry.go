package migrations
 
import (
	"fmt"
	"log"
 
	"gorm.io/gorm"
)

func RunAll(db *gorm.DB) error {
    steps := []struct {
        name string
        fn   func(*gorm.DB) error
    }{
        {"0001_dev1_organizations", Migrate0001Organizations},
        {"0002_dev1_users", Migrate0002Users},
        {"0003_dev1_teams", Migrate0003Teams},
        {"0005_dev3_tasks", MigrateTasks},
        {"0006_dev4_tasks", MigrateTaskAssignments},
        {"0007_dev5_tasks", MigrateTaskTimeLogs},
        {"0100_dev5_tasks_seed", MigrateSeedTasks},
        // future migrations appended here
    }

    for _, step := range steps {
        if err := step.fn(db); err != nil {
            return fmt.Errorf("migration %s failed: %w", step.name, err)
        }
		log.Printf("✅ migration ok: %s", step.name)
    }
    return nil
}
