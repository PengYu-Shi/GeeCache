package router

import (
	"GeeServer/GeeDial"
	"GeeServer/model"
	pb "GeeServer/serverpb"
	"GeeServer/tool"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func Get(c *gin.Context) {
	request := model.Request{}
	if err := c.ShouldBind(&request); err == nil {
		desIp, inerr := tool.HashMap.Get(request.Key)
		if inerr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err})
			return
		}
		request.Type = "Get"
		logrus.Infoln("Will dial to node: ", desIp, " to get key: ", request.Key)
		//TODO: dial to detIp To Get
		// 创建一个通道来接收结果
		resultChan := make(chan *pb.Response, 1)
		errChan := make(chan error, 1)

		// 启动一个goroutine来处理请求
		go func() {
			req, dialerr := GeeDial.Dial(desIp, request)
			if dialerr != nil {
				errChan <- dialerr
				return
			}
			resultChan <- req
		}()

		// 等待结果或错误
		select {
		case req := <-resultChan:
			if req.Err != " " {
				c.JSON(http.StatusBadRequest, gin.H{"err": req.Err})
				return
			}
			c.JSON(http.StatusOK, gin.H{"value": string(req.Value)})
		case dialerr := <-errChan:
			c.JSON(http.StatusBadRequest, gin.H{"err": dialerr})
			return
		}

	}
}

func Add(c *gin.Context) {
	request := model.Request{}
	if err := c.ShouldBind(&request); err == nil {
		desIp, inerr := tool.HashMap.Get(request.Key)
		if inerr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err})
			return
		}
		request.Type = "Get"
		logrus.Infoln("Will dial to node: ", desIp, " to add key: ", request.Key)
		//TODO: dial to detIp To Add

	}
}
