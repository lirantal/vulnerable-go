package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "time"

    "log/slog"
    "html/template"

    // Gin web framework
    "github.com/gin-gonic/gin"

    // SQLite database driver and query builder
	_ "github.com/mattn/go-sqlite3"
    "database/sql"
	"github.com/jmoiron/sqlx"
)

const baseHost = "localtest.me:8080"

// Define custom error variables
var (
    ErrInvalidURL        = fmt.Errorf("invalid URL")
    ErrHTTPRequestFailed = fmt.Errorf("HTTP request failed")
    ErrReadBodyFailed    = fmt.Errorf("failed to read response body")
    ErrJSONUnmarshal     = fmt.Errorf("failed to unmarshal JSON")
    ErrFileDownload      = fmt.Errorf("failed to download file")
    ErrFileWrite         = fmt.Errorf("failed to write file")
    ErrImageResize       = fmt.Errorf("failed to resize image")
)

type FileInfo struct {
    Filename string `json:"filename"`
    Download string `json:"download"`
}

type File struct {
    ID        int       `db:"id"`
    Filename  string    `db:"filename"`
    Signature string    `db:"signature"`
    TenantID  string    `db:"tenant_id"`
    CreatedAt time.Time `db:"created_at"`
}

type UsersResult struct {
    ID          int
    Username    string
    Email       string
}

func downloadAndResize(ctx *gin.Context, tenantID, fileID, fileSize string) error {
    slog.Info("Processing request", "tenantID", tenantID, "fileID", fileID)

    urlStr := fmt.Sprintf("http://%s.%s/storage/%s.json", tenantID, baseHost, fileID)
    slog.Info("Resolved URL", "url", urlStr)

    // Parse the URL to extract the hostname
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        slog.Error("Invalid URL", "error", err)
        return fmt.Errorf("%w: %v", ErrInvalidURL, err)
    }
    slog.Info("Resolved Hostname", "hostname", parsedURL.Hostname())

    // Make HTTP request
    resp, err := http.Get(urlStr)
    if err != nil {
        slog.Error("HTTP request failed", "error", err)
        return fmt.Errorf("%w: %v", ErrHTTPRequestFailed, err)
    }
    defer resp.Body.Close()

    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        slog.Error("Failed to read response body", "error", err)
        return fmt.Errorf("%w: %v", ErrReadBodyFailed, err)
    }

    // Decode JSON data
    var info FileInfo
    err = json.Unmarshal(body, &info)
    if err != nil {
        slog.Error("Failed to unmarshal JSON", "error", err)
        return fmt.Errorf("%w: %v", ErrJSONUnmarshal, err)
    }

    // Download file
    downloadResp, err := http.Get(info.Download)
    if err != nil {
        slog.Error("Failed to download file", "error", err)
        return fmt.Errorf("%w: %v", ErrFileDownload, err)
    }
    defer downloadResp.Body.Close()

    // Create target filename
    targetFilename := fmt.Sprintf("uploads/%s", info.Filename)

    // Read the downloaded file into memory
    fileBytes, err := io.ReadAll(downloadResp.Body)
    if err != nil {
        slog.Error("Failed to read downloaded file", "error", err)
        return fmt.Errorf("%w: %v", ErrReadBodyFailed, err)
    }

    // Save downloaded file
    err = os.WriteFile(targetFilename, fileBytes, 0600)
    if err != nil {
        slog.Error("Failed to write file", "error", err)
        return fmt.Errorf("%w: %v", ErrFileWrite, err)
    }

    convertCmd := fmt.Sprintf("convert %s -resize %sx%s %s", targetFilename, fileSize, fileSize, targetFilename)
    slog.Info("Running command", "command", convertCmd)
    
    _, err = exec.CommandContext(ctx, "sh", "-c", convertCmd).CombinedOutput()
    if err != nil {
        slog.Error("Error resizing image", "error", err)
        return fmt.Errorf("%w: %v", ErrImageResize, err)
    }

    // Record the file operation in the database:
    db, err := sqlx.Open("sqlite3", "./mydb.db")
    defer db.Close()
    if err != nil {
        slog.Error("Failed to open database", "error", err)
        return fmt.Errorf("Failed to open database: %v", err)
    }

    q := "INSERT INTO files (filename, signature, tenant_id, created_at) VALUES ('" + info.Filename + "', '', '" + tenantID + "', '" + time.Now().String() + "')" 
    _, err = db.Exec(q)
    if err != nil {
        slog.Error("Failed to insert record into database", "error", err)
        return fmt.Errorf("Failed to insert record into database: %v", err)
    }

    slog.Info("Downloaded and resized image", "filename", targetFilename)
    return nil
}

func main() {
    // Create a Gin router
    router := gin.Default()
    router.LoadHTMLGlob("templates/*")

    // Nice way to show 2 vulnerabilities in one:
    // 1. SSRF (because the imgURL query parameter is controlled by the user)
    // 1.1. example request: curl "http://localhost:4002/user/image?imgUrl=https://placehold.co/200x200"
    // 2. XSS (because the content-type output is controlled by the remote server and can be set to text/html)
    // 1.2. example request: tbd
    router.GET("/user/image", func(c *gin.Context) {
        var imgURL string
        imgURL = c.Query("imgUrl")

        resp, err := http.Get(imgURL)
        if err != nil {
            slog.Error("Failed to get image", "error", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get image"})
            return
        }
        defer resp.Body.Close()

        c.DataFromReader(http.StatusOK, resp.ContentLength, 
            resp.Header.Get("Content-Type"), resp.Body, nil)
    });


    // This GET route endpoint shows 2 vulnerabilities:
    // 1. Template Injection (SSTI) due to the template compiled with user input
    // 2. XSS due to the query parameter is used in the template
    router.GET("/users", func(c *gin.Context) {
        query := c.Query("q")

        db, err := sql.Open("sqlite3", "mydb.db")
        if err != nil {
            slog.Error("Failed to connect to database", "error", err)
        }
        defer db.Close()
    
        rows, err := db.Query(fmt.Sprintf("SELECT * FROM users WHERE username LIKE '%%%s%%'", query))
        if err != nil {
            c.String(http.StatusInternalServerError, "Error executing query")
            return
        }
        defer rows.Close()
    
        var results []UsersResult
    
        for rows.Next() {
            var id int
            var username string
            var email string
            err := rows.Scan(&id, &username, &email)
            if err != nil {
                slog.Error("Failed to map query results", "error", err)
            }
            results = append(results, UsersResult{id, username, email})
        }
    
        tmpl, err := template.New("search").Parse(fmt.Sprintf(`
                    <h2>Results for username query "%s":</h2>
                    <ul>
                        {{range .Results}}
                            <li><a href="{{.Email}}">{{.Username}}</a></li>
                        {{end}}
                    </ul>
            `, query))
        if err != nil {
            c.String(http.StatusInternalServerError, "Error creating template")
            return
        }
    
        data := struct {
            Query   string
            Results []UsersResult
        }{
            Query:   query,
            Results: results,
        }
    
        // Set the content type to text/html and render the template
        c.Header("Content-Type", "text/html")
    
        err = tmpl.Execute(c.Writer, data)
        if err != nil {
            c.String(http.StatusInternalServerError, "Error executing template")
            return
        }
    })

    router.GET("/cloudpawnery/user", func(c *gin.Context) {
        userId := c.Query("userId")
        redirectPage := c.Query("redirectPage")
        userIds := []string{"1", "2", "3"}

        found := false
        for _, id := range userIds {
            if id == userId {
                found = true
                break
            }
        }
        
        if !found {
            c.HTML(http.StatusOK, "users-not-found.tmpl", gin.H{
                "userId": userId,
            })
            return
        }

        c.HTML(http.StatusOK, "users.tmpl", gin.H{
			"userId": userId,
            "redirectPage": redirectPage,
		})
        return
    })

    router.GET("/welcome", func(c *gin.Context) {
        firstname := c.DefaultQuery("firstname", "Guest")
		lastname := c.Query("lastname")
		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
        return
    })

    // Define a GET endpoint to loop through all the records in the database
    // for the files table and print them (using File struct) and return them
    // as JSON response
    router.GET("/cloudpawnery/image", func(c *gin.Context) {
        // get tenantID from query parameter
        tenantID := c.Query("tenantID")


        db, err := sqlx.Open("sqlite3", "./mydb.db")
        defer db.Close()

        if err != nil {
            slog.Error("Failed to open database", "error", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open database"})
            return
        }

        rows, err := db.Queryx("SELECT * FROM files WHERE tenant_id = '" + tenantID + "'")
        if err != nil {
            slog.Error("Failed to query database", "error", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
            return
        }
        defer rows.Close()

        var files []File
        for rows.Next() {
            var f File
            err := rows.StructScan(&f)
            if err != nil {
                slog.Error("Failed to scan database row", "error", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan database row"})
                return
            }
            files = append(files, f)
        }

        c.JSON(http.StatusOK, gin.H{"files": files})
    })

    // Define a POST endpoint
    router.POST("/cloudpawnery/image", func(c *gin.Context) {
        tenantID := c.Query("tenantID")
        fileID := c.Query("fileID")
        fileSize := c.Query("fileSize")

        if fileSize == "" {
            fileSize = "200"
        }

        // Validate tenantID and fileID
        if tenantID == "" || fileID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Missing tenantID or fileID"})
            return
        }

        // Call the download and resize function
        err := downloadAndResize(c, tenantID, fileID, fileSize)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // Return a success response
        c.JSON(http.StatusOK, gin.H{"message": "File downloaded and resized successfully"})
    })

    // Start the HTTP server
    router.Run(":4002")
}