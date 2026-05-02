package module

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// SubmoduleLoader provides shared functionality for loading submodules and examples
type SubmoduleLoader struct {
	// db provides database access (required)
	db *gorm.DB
}

// NewSubmoduleLoader creates a new submodule loader
// Returns an error if db is nil
func NewSubmoduleLoader(db *gorm.DB) (*SubmoduleLoader, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &SubmoduleLoader{db: db}, nil
}

// LoadSubmodulesAndExamples loads submodules and examples from the database
// and populates them on the given module version.
// The provided db must include the transaction context if called within a transaction.
func (s *SubmoduleLoader) LoadSubmodulesAndExamples(db *gorm.DB, moduleVersion *model.ModuleVersion, moduleVersionID int) error {
	// Load submodules from database (examples are also stored as submodules with type="example")
	var submodulesDB []sqldb.SubmoduleDB
	if err := db.Where("parent_module_version = ?", moduleVersionID).Find(&submodulesDB).Error; err != nil {
		return fmt.Errorf("failed to load submodules: %w", err)
	}

	log.Printf("DEBUG: Found %d submodules for module version %d", len(submodulesDB), moduleVersionID)

	// Convert submodules to domain models and add to module version
	for _, submoduleDB := range submodulesDB {
		// Load module details for submodule if available
		var submoduleDetails *model.ModuleDetails
		if submoduleDB.ModuleDetailsID != nil {
			var detailsDB sqldb.ModuleDetailsDB
			err := db.First(&detailsDB, *submoduleDB.ModuleDetailsID).Error
			if err == nil {
				submoduleDetails = fromDBModuleDetails(&detailsDB)
			}
		}
		if submoduleDetails == nil {
			submoduleDetails = model.NewModuleDetails([]byte{})
		}

		// Determine if this is an example based on type field
		isExample := submoduleDB.Type != nil && *submoduleDB.Type == "example"
		log.Printf("DEBUG: Processing submodule: path=%s, type=%v, isExample=%v", submoduleDB.Path, submoduleDB.Type, isExample)

		if isExample {
			// Create Example and load its files
			example := model.NewExample(
				submoduleDB.Path,
				submoduleDB.Name,
				submoduleDetails,
			)

			// Load example files for this example (submodule)
			var exampleFilesDB []sqldb.ExampleFileDB
			if err := db.Where("submodule_id = ?", submoduleDB.ID).Find(&exampleFilesDB).Error; err != nil {
				return fmt.Errorf("failed to load example files: %w", err)
			}

			log.Printf("DEBUG: Found %d files for example %s", len(exampleFilesDB), submoduleDB.Path)

			// Add files to example
			for _, exampleFileDB := range exampleFilesDB {
				exampleFile := model.NewExampleFile(
					exampleFileDB.Path,
					exampleFileDB.Content,
				)
				example.AddFile(exampleFile)
			}

			moduleVersion.AddExample(example)
			log.Printf("DEBUG: Added example to module version")
		} else {
			// Create Submodule
			submodule := model.NewSubmodule(
				submoduleDB.Path,
				submoduleDB.Name,
				submoduleDB.Type,
				submoduleDetails,
			)
			moduleVersion.AddSubmodule(submodule)
			log.Printf("DEBUG: Added submodule to module version")
		}
	}

	return nil
}
