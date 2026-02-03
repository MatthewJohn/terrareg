package terrareg_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

type authCase struct {
	name     string
	buildReq func(*testing.T, *sqldb.Database, string) *http.Request
}

var authCases = []authCase{
	{
		name: "unauthenticated",
		buildReq: func(t *testing.T, _ *sqldb.Database, path string) *http.Request {
			return testutils.BuildUnauthenticatedRequest(t, http.MethodGet, path)
		},
	},
	{
		name: "authenticated regular user",
		buildReq: func(t *testing.T, db *sqldb.Database, path string) *http.Request {
			req, _ := testutils.BuildAuthenticatedRequestWithSession(
				t,
				db,
				http.MethodGet,
				path,
				"regular-user",
				false,
			)
			return req
		},
	},
	{
		name: "authenticated admin user",
		buildReq: func(t *testing.T, db *sqldb.Database, path string) *http.Request {
			req, _ := testutils.BuildAdminRequest(
				t,
				db,
				http.MethodGet,
				path,
			)
			return req
		},
	},
}

type endpointCase struct {
	name       string
	path       string
	setup      func(*testing.T, *sqldb.Database)
	withChiCtx func(*testing.T, *http.Request) *http.Request
	container  func(*testing.T, *sqldb.Database) *container.Container
}

func TestOptionalAuthEndpoints(t *testing.T) {
	endpoints := []endpointCase{
		{
			name:      "config",
			path:      "/v1/terrareg/config",
			container: testutils.CreateTestContainer,
		},
		{
			name:      "health",
			path:      "/v1/terrareg/health",
			container: testutils.CreateTestContainer,
		},
		{
			name:      "version",
			path:      "/v1/terrareg/version",
			container: testutils.CreateTestContainer,
		},
		{
			name: "module list",
			path: "/v1/terrareg/modules/module-list-namespace",
			setup: func(t *testing.T, db *sqldb.Database) {
				_ = testutils.CreateNamespace(t, db, "module-list-namespace", nil)
			},
			withChiCtx: func(t *testing.T, req *http.Request) *http.Request {
				return testutils.AddChiContext(t, req, map[string]string{
					"namespace": "module-list-namespace",
				})
			},
			container: testutils.CreateTestContainer,
		},
		{
			name: "module versions",
			path: "/v1/terrareg/modules/version-list-namespace/testmod/testprovider/versions",
			setup: func(t *testing.T, db *sqldb.Database) {
				ns := testutils.CreateNamespace(t, db, "version-list-namespace", nil)
				_ = testutils.CreateModuleProvider(t, db, ns.ID, "testmod", "testprovider")
			},
			withChiCtx: func(t *testing.T, req *http.Request) *http.Request {
				return testutils.AddChiContext(t, req, map[string]string{
					"namespace": "version-list-namespace",
					"name":      "testmod",
					"provider":  "testprovider",
				})
			},
			container: func(t *testing.T, db *sqldb.Database) *container.Container {
				return testutils.CreateTestContainerWithConfig(
					t,
					db,
					testutils.WithAllowUnauthenticatedAccess(true),
				)
			},
		},
		{
			name: "analytics",
			path: "/v1/terrareg/analytics/global/stats_summary",
			container: func(t *testing.T, db *sqldb.Database) *container.Container {
				return testutils.CreateTestContainerWithConfig(
					t,
					db,
					testutils.WithAllowUnauthenticatedAccess(true),
				)
			},
		},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			if ep.setup != nil {
				ep.setup(t, db)
			}

			cont := ep.container(t, db)
			router := cont.Server.Router()

			for _, auth := range authCases {
				t.Run(auth.name, func(t *testing.T) {
					req := auth.buildReq(t, db, ep.path)

					if ep.withChiCtx != nil {
						req = ep.withChiCtx(t, req)
					}

					w := testutils.ServeHTTP(router, req)
					assert.Equal(t, http.StatusOK, w.Code)
				})
			}
		})
	}
}
