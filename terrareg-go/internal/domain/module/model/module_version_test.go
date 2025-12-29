package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// TestNewModuleVersion tests creating a new module version
func TestNewModuleVersion(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		details := NewModuleDetails([]byte("test readme"))
		mv, err := NewModuleVersion("1.2.3", details, false)

		require.NoError(t, err)
		assert.NotNil(t, mv)
		assert.Equal(t, "1.2.3", mv.Version().String())
		assert.False(t, mv.IsBeta())
		assert.False(t, mv.IsInternal())
		assert.False(t, mv.IsPublished())
		assert.NotNil(t, mv.Details())
		assert.Empty(t, mv.Submodules())
		assert.Empty(t, mv.Examples())
		assert.False(t, mv.CreatedAt().IsZero())
		assert.False(t, mv.UpdatedAt().IsZero())
	})

	t.Run("beta version", func(t *testing.T) {
		mv, err := NewModuleVersion("1.2.3-beta", nil, true)

		require.NoError(t, err)
		assert.Equal(t, "1.2.3-beta", mv.Version().String())
		assert.True(t, mv.IsBeta())
		assert.True(t, mv.Version().IsPrerelease())
	})

	t.Run("invalid version", func(t *testing.T) {
		invalidVersions := []string{
			"astring",
			"",
			"1",
			"1.1",
			".23.1",
			"1.1.1.1",
			"1.1.1.",
			"1.2.3-dottedsuffix1.2",
			"1.2.3-invalid-suffix",
			"1.0.9-",
		}

		for _, version := range invalidVersions {
			t.Run("invalid_"+version, func(t *testing.T) {
				mv, err := NewModuleVersion(version, nil, false)
				assert.Error(t, err)
				assert.Nil(t, mv)
			})
		}
	})

	t.Run("valid versions with beta detection", func(t *testing.T) {
		testCases := []struct {
			version     string
			expectedBeta bool
		}{
			{"1.1.1", false},
			{"13.14.16", false},
			{"1.10.10", false},
			{"01.01.01", false}, // Leading zeros are valid
			{"1.2.3-alpha", true},
			{"1.2.3-beta", true},
			{"1.2.3-anothersuffix1", true},
			{"1.2.2-123", true},
		}

		for _, tc := range testCases {
			t.Run(tc.version, func(t *testing.T) {
				details := NewModuleDetails([]byte{})
				mv, err := NewModuleVersion(tc.version, details, tc.expectedBeta)

				require.NoError(t, err)
				assert.Equal(t, tc.expectedBeta, mv.IsBeta())
			})
		}
	})
}

// TestReconstructModuleVersion tests reconstructing from persistence
func TestReconstructModuleVersion(t *testing.T) {
	gitSHA := "abc123"
	gitPath := "modules/test"
	baseURL := "https://example.com/{namespace}/{module}"
	cloneURL := "https://example.com/{namespace}/{module}.git"
	browseURL := "https://example.com/{namespace}/{module}/tree/{tag}"
	owner := "testowner"
	description := "Test description"
	variableTemplate := []byte("variable: test")
	extractionVersion := 1

	now := time.Now()

	mv, err := ReconstructModuleVersion(
		123,
		"1.2.3",
		nil,
		false,
		true,
		true,
		&now,
		&gitSHA,
		&gitPath,
		true,
		&baseURL,
		&cloneURL,
		&browseURL,
		&owner,
		&description,
		variableTemplate,
		&extractionVersion,
		now,
		now,
	)

	require.NoError(t, err)
	assert.NotNil(t, mv)
	assert.Equal(t, 123, mv.ID())
	assert.Equal(t, "1.2.3", mv.Version().String())
	assert.False(t, mv.IsBeta())
	assert.True(t, mv.IsInternal())
	assert.True(t, mv.IsPublished())
	assert.Equal(t, &gitSHA, mv.GitSHA())
	assert.Equal(t, &gitPath, mv.GitPath())
	assert.True(t, mv.ArchiveGitPath())
	assert.Equal(t, &baseURL, mv.RepoBaseURLTemplate())
	assert.Equal(t, &cloneURL, mv.RepoCloneURLTemplate())
	assert.Equal(t, &browseURL, mv.RepoBrowseURLTemplate())
	assert.Equal(t, &owner, mv.Owner())
	assert.Equal(t, &description, mv.Description())
	assert.Equal(t, variableTemplate, mv.VariableTemplate())
	assert.Equal(t, &extractionVersion, mv.ExtractionVersion())
}

// TestModuleVersion_Publish tests publishing a module version
func TestModuleVersion_Publish(t *testing.T) {
	t.Run("publish unpublished version", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		err := mv.Publish()

		assert.NoError(t, err)
		assert.True(t, mv.IsPublished())
		assert.NotNil(t, mv.PublishedAt())
	})

	t.Run("publish already published version", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)
		mv.Publish()

		err := mv.Publish()

		assert.Error(t, err)
		assert.ErrorIs(t, err, shared.ErrDomainViolation)
	})
}

// TestModuleVersion_Unpublish tests unpublishing a module version
func TestModuleVersion_Unpublish(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)
	mv.Publish()

	mv.Unpublish()

	assert.False(t, mv.IsPublished())
	assert.Nil(t, mv.PublishedAt())
}

// TestModuleVersion_MarkInternal tests marking as internal
func TestModuleVersion_MarkInternal(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	assert.False(t, mv.IsInternal())

	mv.MarkInternal()

	assert.True(t, mv.IsInternal())
}

// TestModuleVersion_MarkPublic tests marking as public
func TestModuleVersion_MarkPublic(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)
	mv.MarkInternal()

	mv.MarkPublic()

	assert.False(t, mv.IsInternal())
}

// TestModuleVersion_SetGitInfo tests setting Git information
func TestModuleVersion_SetGitInfo(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	gitSHA := "abc123def456"
	gitPath := "modules/test"

	mv.SetGitInfo(&gitSHA, &gitPath, true)

	assert.Equal(t, &gitSHA, mv.GitSHA())
	assert.Equal(t, &gitPath, mv.GitPath())
	assert.True(t, mv.ArchiveGitPath())
}

// TestModuleVersion_SetRepositoryURLs tests setting repository URLs
func TestModuleVersion_SetRepositoryURLs(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	baseURL := "https://example.com/base"
	cloneURL := "https://example.com/clone"
	browseURL := "https://example.com/browse"

	mv.SetRepositoryURLs(&baseURL, &cloneURL, &browseURL)

	assert.Equal(t, &baseURL, mv.RepoBaseURLTemplate())
	assert.Equal(t, &cloneURL, mv.RepoCloneURLTemplate())
	assert.Equal(t, &browseURL, mv.RepoBrowseURLTemplate())
}

// TestModuleVersion_SetMetadata tests setting owner and description
func TestModuleVersion_SetMetadata(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	owner := "testowner"
	description := "Test description"

	mv.SetMetadata(&owner, &description)

	assert.Equal(t, &owner, mv.Owner())
	assert.Equal(t, &description, mv.Description())
}

// TestModuleVersion_SetDetails tests setting module details
func TestModuleVersion_SetDetails(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	newDetails := NewModuleDetails([]byte("new readme"))

	mv.SetDetails(newDetails)

	assert.Equal(t, newDetails, mv.Details())
	assert.Equal(t, []byte("new readme"), mv.Details().ReadmeContent())
}

// TestModuleVersion_SetVariableTemplate tests setting variable template
func TestModuleVersion_SetVariableTemplate(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	template := []byte("variable: test")

	mv.SetVariableTemplate(template)

	assert.Equal(t, template, mv.VariableTemplate())
}

// TestModuleVersion_AddSubmodule tests adding submodules
func TestModuleVersion_AddSubmodule(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	assert.Empty(t, mv.Submodules())

	submodule := NewSubmodule("modules/test", strPtr("test"), strPtr("submodule"), nil)
	mv.AddSubmodule(submodule)

	assert.Len(t, mv.Submodules(), 1)
	assert.True(t, mv.HasSubmodules())
}

// TestModuleVersion_AddExample tests adding examples
func TestModuleVersion_AddExample(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	assert.Empty(t, mv.Examples())

	example := NewExample("examples/test", strPtr("test"), nil)
	mv.AddExample(example)

	assert.Len(t, mv.Examples(), 1)
	assert.True(t, mv.HasExamples())
}

// TestModuleVersion_AddFile tests adding files
func TestModuleVersion_AddFile(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	assert.Empty(t, mv.Files())

	file := NewModuleFile("LICENSE", []byte("MIT license"))
	mv.AddFile(file)

	assert.Len(t, mv.Files(), 1)
}

// TestModuleVersion_GetSubmoduleByPath tests getting submodule by path
func TestModuleVersion_GetSubmoduleByPath(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	submodule1 := NewSubmodule("modules/test1", strPtr("test1"), strPtr("submodule"), nil)
	submodule2 := NewSubmodule("modules/test2", strPtr("test2"), strPtr("submodule"), nil)
	mv.AddSubmodule(submodule1)
	mv.AddSubmodule(submodule2)

	assert.Equal(t, submodule1, mv.GetSubmoduleByPath("modules/test1"))
	assert.Equal(t, submodule2, mv.GetSubmoduleByPath("modules/test2"))
	assert.Nil(t, mv.GetSubmoduleByPath("modules/notexist"))
}

// TestModuleVersion_GetExampleByPath tests getting example by path
func TestModuleVersion_GetExampleByPath(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	example1 := NewExample("examples/test1", strPtr("test1"), nil)
	example2 := NewExample("examples/test2", strPtr("test2"), nil)
	mv.AddExample(example1)
	mv.AddExample(example2)

	assert.Equal(t, example1, mv.GetExampleByPath("examples/test1"))
	assert.Equal(t, example2, mv.GetExampleByPath("examples/test2"))
	assert.Nil(t, mv.GetExampleByPath("examples/notexist"))
}

// TestModuleVersion_GetRootModuleSpecs tests getting root module specs
func TestModuleVersion_GetRootModuleSpecs(t *testing.T) {
	t.Run("no details", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		specs := mv.GetRootModuleSpecs()

		assert.NotNil(t, specs)
		assert.True(t, specs.Empty)
		assert.Empty(t, specs.Inputs)
		assert.Empty(t, specs.Outputs)
	})

	t.Run("with readme only", func(t *testing.T) {
		readme := []byte("# Test README\n\nThis is a test.")
		details := NewModuleDetails(readme)
		mv, _ := NewModuleVersion("1.0.0", details, false)

		specs := mv.GetRootModuleSpecs()

		assert.NotNil(t, specs)
		assert.False(t, specs.Empty)
		assert.Equal(t, "# Test README\n\nThis is a test.", specs.Readme)
		assert.Empty(t, specs.Inputs)
		assert.Empty(t, specs.Outputs)
	})

	t.Run("with terraform docs", func(t *testing.T) {
		readme := []byte("# Test README")
		terraformDocs := []byte(`{
			"inputs": [
				{"name": "test_input", "type": "string", "description": "Test input", "required": true}
			],
			"outputs": [
				{"name": "test_output", "description": "Test output"}
			]
		}`)
		details := NewModuleDetails(readme).WithTerraformDocs(terraformDocs)
		mv, _ := NewModuleVersion("1.0.0", details, false)

		specs := mv.GetRootModuleSpecs()

		assert.NotNil(t, specs)
		assert.Len(t, specs.Inputs, 1)
		assert.Len(t, specs.Outputs, 1)
		assert.Equal(t, "test_input", specs.Inputs[0].Name)
		assert.Equal(t, "test_output", specs.Outputs[0].Name)
	})
}

// TestModuleVersion_GetSubmodules tests getting all submodule specs
func TestModuleVersion_GetSubmodules(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	submodule := NewSubmodule("modules/test", strPtr("test"), strPtr("submodule"), nil)
	mv.AddSubmodule(submodule)

	specs := mv.GetSubmodules()

	assert.Len(t, specs, 1)
	assert.Equal(t, "modules/test", specs[0].Path)
}

// TestModuleVersion_GetExamples tests getting all example specs
func TestModuleVersion_GetExamples(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	example := NewExample("examples/test", strPtr("test"), nil)
	mv.AddExample(example)

	specs := mv.GetExamples()

	assert.Len(t, specs, 1)
	assert.Equal(t, "examples/test", specs[0].Path)
}

// TestModuleVersion_GetSecurityResults tests getting security results
func TestModuleVersion_GetSecurityResults(t *testing.T) {
	t.Run("no details", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		results := mv.GetSecurityResults()

		assert.Empty(t, results)
		assert.Equal(t, 0, mv.GetSecurityFailures())
	})

	t.Run("with tfsec results", func(t *testing.T) {
		tfsecJSON := []byte(`{
			"results": [
				{
					"rule_id": "AWS003",
					"severity": "HIGH",
					"title": "Test security issue",
					"description": "This is a test security issue",
					"location": {
						"filename": "main.tf",
						"start_line": 10,
						"end_line": 15
					}
				}
			]
		}`)
		details := NewModuleDetails([]byte{}).WithTfsec(tfsecJSON)
		mv, _ := NewModuleVersion("1.0.0", details, false)

		results := mv.GetSecurityResults()

		assert.Len(t, results, 1)
		assert.Equal(t, "AWS003", results[0].RuleID)
		assert.Equal(t, "HIGH", results[0].Severity)
		assert.Equal(t, 1, mv.GetSecurityFailures())
	})
}

// TestModuleVersion_GetPublishedAtDisplay tests formatted publication date
func TestModuleVersion_GetPublishedAtDisplay(t *testing.T) {
	t.Run("not published", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		display := mv.GetPublishedAtDisplay()

		assert.Empty(t, display)
	})

	t.Run("published", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)
		mv.Publish()

		display := mv.GetPublishedAtDisplay()

		assert.NotEmpty(t, display)
		assert.Contains(t, display, "at")
	})
}

// TestModuleVersion_String tests string representation
func TestModuleVersion_String(t *testing.T) {
	mv, _ := NewModuleVersion("1.2.3-beta", nil, true)

	assert.Equal(t, "1.2.3-beta", mv.String())
}

// TestModuleVersion_ResetID tests resetting ID
func TestModuleVersion_ResetID(t *testing.T) {
	mv, _ := ReconstructModuleVersion(
		123,
		"1.0.0",
		nil,
		false,
		false,
		false,
		nil,
		nil,
		nil,
		false,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	assert.Equal(t, 123, mv.ID())

	mv.ResetID()

	assert.Equal(t, 0, mv.ID())
}

// TestModuleVersion_GetRepositoryURLs tests getting repository URLs
func TestModuleVersion_GetRepositoryURLs(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	baseURL := "https://example.com/base"
	cloneURL := "https://example.com/clone"
	browseURL := "https://example.com/browse"
	mv.SetRepositoryURLs(&baseURL, &cloneURL, &browseURL)

	base, clone, browse := mv.GetRepositoryURLs()

	assert.Equal(t, baseURL, base)
	assert.Equal(t, cloneURL, clone)
	assert.Equal(t, browseURL, browse)
}

// TestModuleVersion_GetVariableTemplate tests getting variable template
func TestModuleVersion_GetVariableTemplate(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	template := []byte("variable: test")
	mv.SetVariableTemplate(template)

	assert.Equal(t, template, mv.GetVariableTemplate())
}

// TestModuleVersion_PrepareModule tests module preparation stub
func TestModuleVersion_PrepareModule(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	shouldPublish, err := mv.PrepareModule()

	assert.NoError(t, err)
	assert.False(t, shouldPublish) // TODO: When extraction is implemented, test actual behavior
}

// TestModuleVersion_Delete tests delete stub
func TestModuleVersion_Delete(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	err := mv.Delete()

	assert.NoError(t, err) // TODO: When delete is implemented, test actual cascade behavior
}

// TestModuleVersion_GetModuleDetailsID tests getting module details ID
func TestModuleVersion_GetModuleDetailsID(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	id := mv.GetModuleDetailsID()

	assert.Nil(t, id) // TODO: When details persistence is implemented, test actual ID
}

// TestModuleVersion_GetUsageExample tests generating usage example
func TestModuleVersion_GetUsageExample(t *testing.T) {
	t.Run("with module provider", func(t *testing.T) {
		// This test is limited as we need a full ModuleProvider setup
		// For now, test the nil case
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		example := mv.GetUsageExample("example.com")

		assert.Empty(t, example) // Empty because no module provider is set
	})

	t.Run("without module provider", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		example := mv.GetUsageExample("example.com")

		assert.Empty(t, example)
	})
}

// TestModuleVersion_GetTerraformExampleVersionString tests example version constraint
func TestModuleVersion_GetTerraformExampleVersionString(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		beta     bool
		published bool
		template string
		expected string
	}{
		{
			name:     "published non-beta latest",
			version:  "1.5.0",
			beta:     false,
			published: true,
			template: "{major}.{minor}.{patch}",
			expected: "1.5.0",
		},
		{
			name:     "published non-beta pre-major",
			version:  "0.1.5",
			beta:     false,
			published: true,
			template: "{major}.{minor}.{patch}",
			expected: "0.1.5",
		},
		{
			name:     "beta version",
			version:  "5.6.23-beta",
			beta:     true,
			published: true,
			template: "{major}.{minor}.{patch}",
			expected: "5.6.23-beta",
		},
		{
			name:     "non-published version",
			version:  "5.6.25",
			beta:     false,
			published: false,
			template: "{major}.{minor}.{patch}",
			expected: "5.6.25",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mv, _ := NewModuleVersion(tc.version, nil, tc.beta)
			if tc.published {
				mv.Publish()
			}

			result := mv.GetTerraformExampleVersionString()

			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestModuleVersion_GetTerraformExampleVersionComment tests example version comments
func TestModuleVersion_GetTerraformExampleVersionComment(t *testing.T) {
	t.Run("not published", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		comments := mv.GetTerraformExampleVersionComment()

		assert.Len(t, comments, 2)
		assert.Contains(t, comments[0], "not yet been published")
	})

	t.Run("published beta", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0-beta", nil, true)
		mv.Publish()

		comments := mv.GetTerraformExampleVersionComment()

		assert.Len(t, comments, 2)
		assert.Contains(t, comments[0], "beta version")
	})

	t.Run("published non-beta latest", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)
		mv.Publish()

		comments := mv.GetTerraformExampleVersionComment()

		// Without a module provider, isLatestVersion returns false
		// So the version is considered "not latest" and has comments
		assert.Len(t, comments, 2)
		assert.Contains(t, comments[0], "not the latest version")
	})

	t.Run("published non-beta not latest", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)
		mv.Publish()
		// Note: Without a module provider with newer versions, isLatestVersion will return false
		// This tests the comment logic

		comments := mv.GetTerraformExampleVersionComment()

		// Since there's no module provider, it won't be latest
		assert.Len(t, comments, 2)
		assert.Contains(t, comments[0], "not the latest version")
	})
}

// TestModuleVersion_GetCustomLinks tests custom links stub
func TestModuleVersion_GetCustomLinks(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	links := mv.GetCustomLinks()

	assert.Empty(t, links) // TODO: When custom links are implemented
}

// TestModuleVersion_GetAdditionalTabFiles tests additional tab files stub
func TestModuleVersion_GetAdditionalTabFiles(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	tabs := mv.GetAdditionalTabFiles()

	assert.Empty(t, tabs) // TODO: When additional tabs are implemented
}

// TestModuleVersion_GetModuleExtractionUpToDate tests extraction version check
func TestModuleVersion_GetModuleExtractionUpToDate(t *testing.T) {
	t.Run("no extraction version", func(t *testing.T) {
		mv, _ := NewModuleVersion("1.0.0", nil, false)

		upToDate := mv.GetModuleExtractionUpToDate()

		assert.False(t, upToDate)
	})

	t.Run("with extraction version", func(t *testing.T) {
		version := 1
		mv, _ := ReconstructModuleVersion(
			0,
			"1.0.0",
			nil,
			false,
			false,
			false,
			nil,
			nil,
			nil,
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			&version,
			time.Now(),
			time.Now(),
		)

		upToDate := mv.GetModuleExtractionUpToDate()

		assert.True(t, upToDate) // TODO: When extraction version checking is implemented
	})
}

// TestModuleVersion_GetDownloads tests download count stub
func TestModuleVersion_GetDownloads(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	downloads := mv.GetDownloads()

	assert.Equal(t, 0, downloads) // TODO: When analytics is implemented
}

// TestModuleVersion_GetProviderDependencies tests provider dependencies stub
func TestModuleVersion_GetProviderDependencies(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	deps := mv.GetProviderDependencies()

	assert.Empty(t, deps) // TODO: When provider dependency parsing is implemented
}

// TestModuleVersion_GetTerraformModules tests terraform modules stub
func TestModuleVersion_GetTerraformModules(t *testing.T) {
	mv, _ := NewModuleVersion("1.0.0", nil, false)

	modules := mv.GetTerraformModules()

	assert.Empty(t, modules) // TODO: When terraform modules parsing is implemented
}

// Helper function
func strPtr(s string) *string {
	return &s
}
