package migrations
 
import (
	"fmt"
 
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
        // future migrations appended here
    }

    for _, step := range steps {
        if err := step.fn(db); err != nil {
            return fmt.Errorf("migration %s failed: %w", step.name, err)
        }
    }
    return nil
}
