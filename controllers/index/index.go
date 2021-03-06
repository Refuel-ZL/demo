package index

import (
	"demo/lib/ws"
	jwtauth "demo/middleware/JWT"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type wsinfo struct {
	wsConn *websocket.Conn
	conn   *ws.Connection
}

type msg struct {
	Id      string
	Type    string
	Content string
}

var wslist map[string]wsinfo = make(map[string]wsinfo)

// func1: 处理最基本的GET
func Index(c *gin.Context) {
	// 回复一个200OK,在client的http-get的resp的body中获取数据
	c.String(http.StatusOK, "index")
}
func Jwt(c *gin.Context) {
	j := &jwtauth.JWT{
		[]byte("test"),
	}
	claims := jwtauth.CustomClaims{
		1,
		"awh521",
		"1044176017@qq.com",
		jwt.StandardClaims{
			ExpiresAt: 15000, //time.Now().Add(24 * time.Hour).Unix()
			Issuer:    "test",
		},
	}
	token, err := j.CreateToken(claims)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		c.Abort()
	}
	c.String(http.StatusOK, token+"---------------<br>")
	res, err := j.ParseToken(token)
	if err != nil {
		if err == jwtauth.TokenExpired {
			newToken, err := j.RefreshToken(token)
			if err != nil {
				c.String(http.StatusOK, err.Error())
			} else {
				c.String(http.StatusOK, newToken)
			}
		} else {
			c.String(http.StatusOK, err.Error())
		}
	} else {
		c.JSON(http.StatusOK, res)
	}
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func Ws(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		return
	}
	var (
		wsConn *websocket.Conn
		err    error
		data   []byte
		conn   *ws.Connection
		info   *wsinfo
	)

	if wsConn, err = upgrader.Upgrade(c.Writer, c.Request, nil); err != nil {
		return
	}
	if conn, err = ws.InitConnection(wsConn); err != nil {
		goto ERR
	}
	info = &wsinfo{
		wsConn: wsConn,
		conn:   conn,
	}
	wslist[id] = *info

	go func() {
		var err error
		for {
			if err = conn.WriteMessage([]byte("心跳包")); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		} else {
			fmt.Println("收到：" + string(data))
			for wsid := range wslist {
				var c []byte = nil
				if id != wsid {
					mg := &msg{
						Id:      id,
						Type:    "text",
						Content: id + `:发送` + string(data),
					}
					c, _ = json.Marshal(mg)
				} else {
					mg := &msg{
						Id:      "server",
						Type:    "text",
						Content: `已收到` + string(data),
					}
					c, _ = json.Marshal(mg)
				}
				if err = wslist[wsid].conn.WriteMessage(c); err != nil {
					goto ERR
				}
			}
		}
	}
ERR:
	delete(wslist, id)
	wslist[id].conn.Close()
}
