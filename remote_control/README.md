### Recv file with Gin

```go
import (
	"github.com/gin-gonic/gin"
)

func Recv(c *gin.Context) {
	file, _ := c.FormFile("file")
	dst := c.Request.Header.Get("target")
	c.SaveUploadedFile(file, dst)
	c.JSON(200, gin.H{
		"ok": 1,
	})
}
```