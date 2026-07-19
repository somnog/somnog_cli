package cmd

var Version = "1.0.0"

const (
    DefaultTemplateRepo = "https://github.com/somnog/somnog.git"
    ProjectConfigFile   = "somnog.config.ts"
    BinaryName          = "somnog"

    DefaultPageSize    = 20
    MaxPageSize        = 100
    DefaultSortKey     = "created_at"
    DefaultSortDir     = "desc"

    MarkerHandlers        = "// somnog:handlers"
    MarkerRoutesProtected = "// somnog:routes:protected"
    MarkerRoutesAdmin     = "// somnog:routes:admin"
    MarkerTypes           = "// somnog:types"
    MarkerSchemas         = "// somnog:schemas"
    MarkerResources       = "// somnog:resources"
    MarkerResourceList    = "// somnog:resource-list"

    APIRoutesPath = "apps/api/internal/routes/routes.go"
    ModelsPath    = "apps/api/internal/models"
    ServicesPath  = "apps/api/internal/services"
    HandlersPath  = "apps/api/internal/handlers"
    TSTypesPath   = "packages/shared/types"
    TSSchemasPath = "packages/shared/schemas"
    AdminResPath  = "apps/admin/resources"
)