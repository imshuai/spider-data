package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"

	_ "spider-data/statik"

	"github.com/rakyll/statik/fs"
)

var (
	storage *bolt.DB
	c       *config
	err     error
)

type imgccc struct {
	Name         string `json:"name"`
	LastModified string `json:"last-modified"`
	URL          string `json:"url"`
	Size         string `json:"size"`
}

func (i *imgccc) Dump() string {
	byts, _ := json.Marshal(i)
	return string(byts)
}

func init() {
	c = &config{}
	if len(os.Args) == 1 {
		c.Init("settings.json")
	} else {
		c.Init(os.Args[1])
	}
	storage, err = bolt.Open(c.DBPath, os.ModePerm, nil)
	if err != nil {
		fmt.Printf("load database fail with error:[%v], exit!\n", err)
		os.Exit(1)
	}
}

func main() {

	fs, _ := fs.New()

	gin.SetMode(gin.ReleaseMode)

	e := gin.New()
	e.Use(gin.Recovery(), gin.Logger())

	e.Use(func() gin.HandlerFunc {
		fileserver := http.FileServer(fs)
		return func(c *gin.Context) {
			if _, err := fs.Open(c.Request.URL.Path); err == nil {
				fileserver.ServeHTTP(c.Writer, c.Request)
				c.Abort()
			}
		}
	}())

	e.POST("/api/:project/:id", func(ctx *gin.Context) {
		project := ctx.Param("project")
		id := ctx.Param("id")
		key := project + ":data:" + id
		data := &imgccc{}
		ctx.BindJSON(data)
		err = storage.Update(func(t *bolt.Tx) error {
			bck, err := t.CreateBucketIfNotExists([]byte(project))
			if err != nil {
				return err
			}
			err = bck.Put([]byte(key), []byte(data.Dump()))
			if err != nil {
				return err
			}
			if total := bck.Get([]byte(project + ":total")); total == nil {
				bck.Put([]byte(project+":total"), []byte("0"))
			} else {
				bck.Put([]byte(project+":total"), itob(btoi(total)+1))
			}
			return nil
		})
		if err != nil {
			ctx.String(200, "%s", "fail")
			return
		}
		ctx.String(200, "%s", "ok")
	})

	e.DELETE("/api/:project/:id", func(ctx *gin.Context) {
		project := ctx.Param("project")
		id := ctx.Param("id")
		key := project + ":data:" + id
		err = storage.Update(func(t *bolt.Tx) error {
			bck, err := t.CreateBucketIfNotExists([]byte(project))
			if err != nil {
				return err
			}
			bck.Delete([]byte(key))
			if total := bck.Get([]byte(project + ":total")); total == nil || btoi(total) < 2 {
				bck.Put([]byte(project+":total"), []byte("0"))
			} else {
				bck.Put([]byte(project+":total"), itob(btoi(total)-1))
			}
			return nil
		})
		ctx.JSON(200, gin.H{
			"status": "成功",
		})
	})

	e.GET("/api/:project", func(ctx *gin.Context) {
		var err error
		var limit int
		key := ""
		act := "next"
		var total int
		project := ctx.Param("project")
		key, _ = ctx.GetQuery("key")
		act, _ = ctx.GetQuery("act")
		n, exist := ctx.GetQuery("num")
		if !exist {
			limit = 20
		} else {
			limit, err = strconv.Atoi(n)
			if err != nil {
				fmt.Println(err)
				ctx.Status(400)
				return
			}
		}

		var vals []map[string][]byte
		err = storage.View(func(t *bolt.Tx) error {
			bck := t.Bucket([]byte(project))
			if bck == nil {
				return fmt.Errorf("BUCKET:[%s] does not exist", project)
			}
			c := bck.Cursor()
			var k, v []byte
			if key == "" || bck.Get([]byte(key)) == nil {
				key = project + ":data:"
			}
			k, v = c.Seek([]byte(key))
			for i := 0; i < limit; i++ {
				val := make(map[string][]byte)
				switch act {
				case "prev":
					k, v = c.Prev()
					if k == nil {
						c.Seek([]byte(key))
						c.Prev()
						act = "next"
						continue
					}
					val[string(k)] = v
					vals = append([]map[string][]byte{val}, vals...)
				case "mount":
					val[string(k)] = v
					vals = append(vals, val)
					act = "next"
				default:
					k, v = c.Next()
					if k == nil {
						break
					}
					val[string(k)] = v
					vals = append(vals, val)
				}

			}
			if t := bck.Get([]byte(project + ":total")); t != nil {
				total = btoi(t)
			}
			return nil
		})
		if err != nil {
			fmt.Println(err)
			ctx.Status(400)
			return
		}
		var data []*imgccc
		for _, v := range vals {
			for _, vv := range v {
				t := &imgccc{}
				json.Unmarshal(vv, t)
				if t.Name != "" {
					data = append(data, t)
				}
			}
		}
		ctx.JSON(200, gin.H{
			"data":  data,
			"total": total,
		})
	})

	e.Run(c.ListenAddress + ":" + c.ListenPort)
}

func itob(v int) []byte {
	vv := strconv.Itoa(v)
	return []byte(vv)
}

func btoi(v []byte) int {
	vv, _ := strconv.Atoi(string(v))
	return vv
}
