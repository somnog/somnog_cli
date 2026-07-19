package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate code for your project",
}

var generateResourceCmd = &cobra.Command{
    Use:   "resource [Name]",
    Short: "Generate a full resource (model, handler, service, routes, schema, types, admin page)",
    Long: `Generates the complete vertical slice for a new resource:

  Backend (Go):
    • apps/api/internal/models/<name>.go       — GORM model with UUID + timestamps
    • apps/api/internal/services/<name>_service.go — CRUD service layer
    • apps/api/internal/handlers/<name>_handler.go — Gin HTTP handler
    • apps/api/internal/routes/routes.go       — Routes injected at marker

  Frontend (TypeScript):
    • packages/shared/types/<name>.ts          — TypeScript interface
    • packages/shared/schemas/<name>.ts        — Zod validation schemas
    • packages/shared/types/index.ts           — Export updated
    • packages/shared/schemas/index.ts         — Export updated
    • apps/admin/resources/<name>.ts           — Admin panel resource config
    • apps/admin/resources/index.ts            — Export updated

Example:
  somnog generate resource Invoice
  somnog generate resource ProductCategory`,
    Args: cobra.ExactArgs(1),
    Run:  runGenerateResource,
}

func init() {
    generateCmd.AddCommand(generateResourceCmd)
}

func runGenerateResource(cmd *cobra.Command, args []string) {
    name := args[0]

    if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
        fmt.Fprintln(os.Stderr, "Error: resource name must start with an uppercase letter (e.g., Invoice, ProductCategory)")
        os.Exit(1)
    }

    root := findProjectRoot()

    snake := toSnake(name)
    plural := toPlural(name)
    lower := toLower(name)
    camel := toCamel(name)

    fmt.Printf("Generating resource: %s\n", name)
    fmt.Println()

    goModelPath := filepath.Join(root, "apps", "api", "internal", "models", snake+".go")
    writeFile(goModelPath, goModelTemplate(name, snake, plural, lower))
    fmt.Printf("  ✓  %s\n", relPath(root, goModelPath))

    goSvcPath := filepath.Join(root, "apps", "api", "internal", "services", snake+"_service.go")
    writeFile(goSvcPath, goServiceTemplate(name, snake, plural, lower))
    fmt.Printf("  ✓  %s\n", relPath(root, goSvcPath))

    goHandlerPath := filepath.Join(root, "apps", "api", "internal", "handlers", snake+"_handler.go")
    writeFile(goHandlerPath, goHandlerTemplate(name, snake, plural, lower, camel))
    fmt.Printf("  ✓  %s\n", relPath(root, goHandlerPath))

    routesPath := filepath.Join(root, "apps", "api", "internal", "routes", "routes.go")
    injectRoutes(routesPath, name, snake, plural, lower, camel)
    fmt.Printf("  ✓  %s  (updated)\n", relPath(root, routesPath))

    tsTypePath := filepath.Join(root, "packages", "shared", "types", lower+".ts")
    writeFile(tsTypePath, tsTypeTemplate(name))
    fmt.Printf("  ✓  %s\n", relPath(root, tsTypePath))

    tsSchemaPath := filepath.Join(root, "packages", "shared", "schemas", lower+".ts")
    writeFile(tsSchemaPath, tsSchemaTemplate(name))
    fmt.Printf("  ✓  %s\n", relPath(root, tsSchemaPath))

    tsTypesIndex := filepath.Join(root, "packages", "shared", "types", "index.ts")
    injectAtMarker(tsTypesIndex, "// somnog:types",
        fmt.Sprintf("export type { %s } from \"./%s\";\n// somnog:types", name, lower))
    fmt.Printf("  ✓  %s  (updated)\n", relPath(root, tsTypesIndex))

    tsSchemasIndex := filepath.Join(root, "packages", "shared", "schemas", "index.ts")
    injectAtMarker(tsSchemasIndex, "// somnog:schemas",
        fmt.Sprintf("export {\n  %sSchema,\n  Create%sSchema,\n  Update%sSchema,\n  type Create%sInput,\n  type Update%sInput,\n} from \"./%s\";\n// somnog:schemas",
            name, name, name, name, name, lower))
    fmt.Printf("  ✓  %s  (updated)\n", relPath(root, tsSchemasIndex))

    adminResPath := filepath.Join(root, "apps", "admin", "resources", lower+".ts")
    writeFile(adminResPath, adminResourceTemplate(name, plural, lower))
    fmt.Printf("  ✓  %s\n", relPath(root, adminResPath))

    adminResIndex := filepath.Join(root, "apps", "admin", "resources", "index.ts")
    injectAdminResource(adminResIndex, name, lower)
    fmt.Printf("  ✓  %s  (updated)\n", relPath(root, adminResIndex))

    fmt.Println()
    fmt.Printf("✅  Resource \"%s\" generated successfully!\n", name)
    fmt.Println()
    fmt.Println("Next steps:")
    fmt.Printf("  1. Edit the model fields in apps/api/internal/models/%s.go\n", snake)
    fmt.Printf("  2. Run: somnog migrate\n")
    fmt.Printf("  3. Update the admin resource columns in apps/admin/resources/%s.ts\n", lower)
}

func goModelTemplate(name, snake, plural, lower string) string {
    _ = plural
    _ = lower
    return fmt.Sprintf(`package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type %s struct {
    ID        string         `+"`"+`gorm:"primarykey;size:36" json:"id"`+"`"+`
    Name      string         `+"`"+`gorm:"size:255;not null" json:"name" binding:"required"`+"`"+`
    CreatedAt time.Time      `+"`"+`json:"created_at"`+"`"+`
    UpdatedAt time.Time      `+"`"+`json:"updated_at"`+"`"+`
    DeletedAt gorm.DeletedAt `+"`"+`gorm:"index" json:"-"`+"`"+`
}

func (m *%s) BeforeCreate(tx *gorm.DB) error {
    if m.ID == "" {
        m.ID = uuid.New().String()
    }
    return nil
}
`, name, name)
}

func goServiceTemplate(name, snake, plural, lower string) string {
    _ = snake
    return fmt.Sprintf(`package services

import (
    "fmt"
    "math"

    "gorm.io/gorm"

    "somnog/apps/api/internal/models"
)

type %sService struct {
    DB *gorm.DB
}

func New%sService(db *gorm.DB) *%sService {
    return &%sService{DB: db}
}

func (s *%sService) List(page, pageSize int, search, sortKey, sortDir string) ([]models.%s, int64, int, error) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    if sortDir != "asc" && sortDir != "desc" {
        sortDir = "desc"
    }
    if sortKey == "" {
        sortKey = "created_at"
    }

    query := s.DB.Model(&models.%s{})

    if search != "" {
        query = query.Where("name ILIKE ?", "%%"+search+"%%")
    }

    var total int64
    query.Count(&total)

    var items []models.%s
    offset := (page - 1) * pageSize
    if err := query.Order(sortKey + " " + sortDir).Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
        return nil, 0, 0, fmt.Errorf("fetching %s: %%w", err)
    }

    pages := int(math.Ceil(float64(total) / float64(pageSize)))
    return items, total, pages, nil
}

func (s *%sService) GetByID(id string) (*models.%s, error) {
    var item models.%s
    if err := s.DB.First(&item, "id = ?", id).Error; err != nil {
        return nil, fmt.Errorf("%s not found: %%w", err)
    }
    return &item, nil
}

func (s *%sService) Create(item *models.%s) error {
    if err := s.DB.Create(item).Error; err != nil {
        return fmt.Errorf("creating %s: %%w", err)
    }
    return nil
}

func (s *%sService) Update(id string, data map[string]interface{}) (*models.%s, error) {
    var item models.%s
    if err := s.DB.First(&item, "id = ?", id).Error; err != nil {
        return nil, fmt.Errorf("%s not found: %%w", err)
    }
    if err := s.DB.Model(&item).Updates(data).Error; err != nil {
        return nil, fmt.Errorf("updating %s: %%w", err)
    }
    s.DB.First(&item, "id = ?", id)
    return &item, nil
}

func (s *%sService) Delete(id string) error {
    var item models.%s
    if err := s.DB.First(&item, "id = ?", id).Error; err != nil {
        return fmt.Errorf("%s not found: %%w", err)
    }
    if err := s.DB.Delete(&item).Error; err != nil {
        return fmt.Errorf("deleting %s: %%w", err)
    }
    return nil
}
`,
        name,
        name, name, name,
        name, name,
        name,
        name,
        lower,
        name, name, name,
        lower,
        name, name,
        lower,
        name, name, name,
        lower,
        lower,
        name, name,
        lower,
        lower,
    )
}

func goHandlerTemplate(name, snake, plural, lower, camel string) string {
    _ = snake
    _ = camel
    return fmt.Sprintf(`package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "somnog/apps/api/internal/models"
    "somnog/apps/api/internal/services"
)

type %sHandler struct {
    DB      *gorm.DB
    Service *services.%sService
}

func New%sHandler(db *gorm.DB) *%sHandler {
    return &%sHandler{
        DB:      db,
        Service: services.New%sService(db),
    }
}

func (h *%sHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    search := c.Query("search")
    sortBy := c.DefaultQuery("sort_by", "created_at")
    sortOrder := c.DefaultQuery("sort_order", "desc")

    items, total, pages, err := h.Service.List(page, pageSize, search, sortBy, sortOrder)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to fetch %s"},
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data": items,
        "meta": gin.H{"total": total, "page": page, "page_size": pageSize, "pages": pages},
    })
}

func (h *%sHandler) GetByID(c *gin.Context) {
    id := c.Param("id")
    item, err := h.Service.GetByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": gin.H{"code": "NOT_FOUND", "message": "%s not found"},
        })
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *%sHandler) Create(c *gin.Context) {
    var req struct {
        Name string `+"`"+`json:"name" binding:"required"`+"`"+`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{
            "error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
        })
        return
    }

    item := models.%s{Name: req.Name}
    if err := h.Service.Create(&item); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create %s"},
        })
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": item, "message": "%s created successfully"})
}

func (h *%sHandler) Update(c *gin.Context) {
    id := c.Param("id")
    var req struct {
        Name string `+"`"+`json:"name"`+"`"+`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{
            "error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
        })
        return
    }

    updates := map[string]interface{}{}
    if req.Name != "" {
        updates["name"] = req.Name
    }

    item, err := h.Service.Update(id, updates)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update %s"},
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": item, "message": "%s updated successfully"})
}

func (h *%sHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.Service.Delete(id); err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": gin.H{"code": "NOT_FOUND", "message": "%s not found"},
        })
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "%s deleted successfully"})
}
`,
        name, name,
        name, name, name, name,
        name,
        plural,
        name,
        name,
        name,
        name,
        lower,
        name,
        name,
        lower,
        name,
        lower,
        name,
        name, // %s not found
        name, // %s deleted successfully
    )
}

func tsTypeTemplate(name string) string {
    return fmt.Sprintf(`export interface %s {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}
`, name)
}

func tsSchemaTemplate(name string) string {
    return fmt.Sprintf(`import { z } from "zod";

export const %sSchema = z.object({
  id: z.string(),
  name: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const Create%sSchema = z.object({
  name: z.string().min(1, "Name is required"),
});

export const Update%sSchema = Create%sSchema.partial();

export type Create%sInput = z.infer<typeof Create%sSchema>;
export type Update%sInput = z.infer<typeof Update%sSchema>;
`, name, name, name, name, name, name, name, name)
}

func adminResourceTemplate(name, plural, lower string) string {
    return fmt.Sprintf(`import { defineResource } from "@/lib/resource";

export const %sResource = defineResource({
  name: "%s",
  slug: "%s",
  endpoint: "/api/admin/%s",
  icon: "Box",
  label: { singular: "%s", plural: "%s" },

  table: {
    columns: [
      { key: "name", label: "Name", sortable: true, searchable: true },
      { key: "created_at", label: "Created", format: "relative", sortable: true },
    ],
    searchable: true,
    searchPlaceholder: "Search %s...",
    actions: ["create", "view", "edit", "delete"],
    bulkActions: ["delete"],
    defaultSort: { key: "created_at", direction: "desc" },
    pageSize: 20,
  },

  form: {
    layout: "single",
    fields: [
      {
        key: "name",
        label: "Name",
        type: "text",
        required: true,
        placeholder: "Enter %s name",
      },
    ],
  },
});
`, toCamel(name), name, lower, plural, name, plural, plural, lower)
}

func injectRoutes(path, name, snake, plural, lower, camel string) {
    _ = camel
    data, err := os.ReadFile(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading routes file: %v\n", err)
        os.Exit(1)
    }
    content := string(data)

    handlerDecl := fmt.Sprintf("\t%sHandler := New%sHandler(db)\n\t// somnog:handlers", lower+"Handler", name)
    content = strings.Replace(content, "\t// somnog:handlers", handlerDecl, 1)

    protectedRoutes := fmt.Sprintf(
        "\t\tprotected.GET(\"/%s\", %sHandler.List)\n\t\tprotected.GET(\"/%s/:id\", %sHandler.GetByID)\n\t\t// somnog:routes:protected",
        plural, lower+"Handler",
        plural, lower+"Handler",
    )
    content = strings.Replace(content, "\t\t// somnog:routes:protected", protectedRoutes, 1)

    adminRoutes := fmt.Sprintf(
        "\t\tadmin.GET(\"/admin/%s\", %sHandler.List)\n\t\tadmin.GET(\"/admin/%s/:id\", %sHandler.GetByID)\n\t\tadmin.POST(\"/admin/%s\", %sHandler.Create)\n\t\tadmin.PUT(\"/admin/%s/:id\", %sHandler.Update)\n\t\tadmin.DELETE(\"/admin/%s/:id\", %sHandler.Delete)\n\t\t// somnog:routes:admin",
        plural, lower+"Handler",
        plural, lower+"Handler",
        plural, lower+"Handler",
        plural, lower+"Handler",
        plural, lower+"Handler",
    )
    content = strings.Replace(content, "\t\t// somnog:routes:admin", adminRoutes, 1)

    writeFileContent(path, content)
}

func injectAtMarker(path, marker, replacement string) {
    data, err := os.ReadFile(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", path, err)
        os.Exit(1)
    }
    content := string(data)
    if !strings.Contains(content, marker) {
        fmt.Fprintf(os.Stderr, "Warning: marker %q not found in %s\n", marker, path)
        return
    }
    content = strings.Replace(content, marker, replacement, 1)
    writeFileContent(path, content)
}

func injectAdminResource(path, name, lower string) {
    data, err := os.ReadFile(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", path, err)
        os.Exit(1)
    }
    content := string(data)

    importLine := fmt.Sprintf("import { %sResource } from \"./%s\";\n// somnog:resources", toCamel(name), lower)
    content = strings.Replace(content, "// somnog:resources", importLine, 1)

    listEntry := fmt.Sprintf("  %sResource,\n  // somnog:resource-list", toCamel(name))
    content = strings.Replace(content, "  // somnog:resource-list", listEntry, 1)

    writeFileContent(path, content)
}

func writeFile(path, content string) {
    if _, err := os.Stat(path); err == nil {
        fmt.Fprintf(os.Stderr, "Error: file already exists: %s\n", path)
        fmt.Fprintln(os.Stderr, "Delete it first or choose a different resource name.")
        os.Exit(1)
    }
    writeFileContent(path, content)
}

func writeFileContent(path, content string) {
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
        os.Exit(1)
    }
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", path, err)
        os.Exit(1)
    }
}

func relPath(root, path string) string {
    rel, err := filepath.Rel(root, path)
    if err != nil {
        return path
    }
    return rel
}