package aop

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"time"
)

func Logrus() gin.HandlerFunc {
	//log := logrus.New()
	//path := viper.GetString("logging.path")
	//lvl := viper.GetString("logging.level")
	//output := viper.Get("logging.out_put")
	//maxAge := viper.GetInt("logging.max_age")
	//rotationTime := viper.GetInt("logging.rotation")
	//level, err := logrus.ParseLevel(lvl)
	//if err != nil && lvl == "" {
	//	log.SetLevel(logrus.InfoLevel)
	//} else {
	//	log.SetLevel(level)
	//}
	//log.SetReportCaller(true)
	//log.SetFormatter(new(logger.MineFormatter))
	//writer, _ := rotatelogs.New(
	//	path+"%Y%m%d%H%M.log",
	//	rotatelogs.WithLinkName(path),
	//	rotatelogs.WithMaxAge(time.Duration(maxAge)*time.Second),
	//	rotatelogs.WithRotationTime(time.Duration(rotationTime)*time.Second),
	//)
	//switch output {
	//case "file":
	//case "fileAndStd":
	//	writers := []io.Writer{writer, os.Stdout}
	//	fileAndStdoutWriter := io.MultiWriter(writers...)
	//	log.SetOutput(fileAndStdoutWriter)
	//case "std":
	//	log.SetOutput(writer)
	//default:
	//	writers := []io.Writer{writer, os.Stdout}
	//	fileAndStdoutWriter := io.MultiWriter(writers...)
	//	log.SetOutput(fileAndStdoutWriter)
	//}
	//
	//writeMap := lfshook.WriterMap{
	//	logrus.InfoLevel:  writer,
	//	logrus.FatalLevel: writer,
	//	logrus.DebugLevel: writer,
	//	logrus.WarnLevel:  writer,
	//	logrus.ErrorLevel: writer,
	//	logrus.PanicLevel: writer,
	//}
	//
	//log.AddHook(lfshook.NewHook(writeMap, &logrus.JSONFormatter{
	//	TimestampFormat: "2006-01-02 15:04:05",
	//}))

	return func(c *gin.Context) {
		//开始时间
		startTime := time.Now()
		//处理请求
		c.Next()
		//结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		//请求方式
		reqMethod := c.Request.Method
		//请求路由
		reqUrl := c.Request.RequestURI
		//状态码
		statusCode := c.Writer.Status()
		//请求ip
		clientIP := c.ClientIP()

		// 日志格式
		fields := logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUrl,
		}
		logrus.WithFields(fields).Info()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录请求结束时间
		duration := time.Since(start)

		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 打印日志
		logger.GetLogger().Infof("Method: %s | Path: %s | Status: %d | Duration: %v | IP: %s\n",
			method, path, statusCode, duration, clientIP)
	}
}
