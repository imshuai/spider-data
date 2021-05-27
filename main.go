package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	_ "spider-data/statik"

	"github.com/rakyll/statik/fs"
)

var (
	rpool *redis.Pool
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
	rpool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379",
				redis.DialConnectTimeout(time.Second*6),
				redis.DialDatabase(0),
				redis.DialKeepAlive(time.Second*10))
		},
		MaxIdle:   0,
		MaxActive: 10,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) > time.Minute {
				_, err := c.Do("PING")
				return err
			}
			return nil
		},
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
		key := project + ":" + id
		score := time.Now().Unix()
		data := &imgccc{}
		ctx.BindJSON(data)
		conn := rpool.Get()
		defer conn.Close()
		conn.Do("ZADD", project, score, id)
		conn.Do("SET", key, data.Dump())
		ctx.String(200, "%s", "ok")
	})

	e.DELETE("/api/:project/:id", func(ctx *gin.Context) {
		project := ctx.Param("project")
		id := ctx.Param("id")
		key := project + ":" + id
		conn := rpool.Get()
		defer conn.Close()
		conn.Do("ZREM", project, id)
		conn.Do("DEL", key)
		ctx.JSON(200, gin.H{
			"status": "成功",
		})
	})

	e.GET("/api/:project/:page", func(ctx *gin.Context) {
		var err error
		var page int
		var limit int
		project := ctx.Param("project")
		page, err = strconv.Atoi(ctx.Param("page"))
		if err != nil {
			ctx.Status(400)
			return
		}
		n, exist := ctx.GetQuery("num")
		if !exist {
			limit = 20
		} else {
			limit, err = strconv.Atoi(n)
			if err != nil {
				ctx.Status(400)
				return
			}
		}
		conn := rpool.Get()
		defer conn.Close()
		if conn.Err() != nil {
			ctx.Status(500)
			return
		}
		total, _ := redis.Int(conn.Do("ZCARD", project))
		keys, _ := redis.Strings(conn.Do("ZRANGE", project, (page-1)*limit, page*limit-1))
		if len(keys) == 0 {
			ctx.Redirect(304, "/api/"+project+"/1")
			return
		}
		var data []imgccc
		for _, key := range keys {
			s, _ := redis.String(conn.Do("GET", project+":"+key))
			i := &imgccc{}
			json.Unmarshal([]byte(s), i)
			data = append(data, *i)
		}
		ctx.JSON(200, gin.H{
			"data":  data,
			"total": total,
		})
	})

	e.Run(os.Args[1])
}
