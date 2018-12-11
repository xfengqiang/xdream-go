package logger

import (
	"testing"
	"encoding/json"
	"log"
	"time"
	"go.uber.org/zap"
	"fmt"
	"github.com/jonboulle/clockwork"
	"os"
)

func setupLogger(cfgStr string, t *testing.T)  {
	var cfg LogConfig
	err := json.Unmarshal([]byte(cfgStr), &cfg)
	if err != nil {
		t.Error("json decode error:"+err.Error())
	}

	b, _ := json.MarshalIndent(cfg, "abc", "\t")
	log.Println(string(b))
	InitLogger(&cfg)
}

func TestLogWrite(t *testing.T)  {
var cfgStr string = `{
	"common_field":{
		"app":"log-demo"
	},
	"log_dir":"./log",
	"routes":[
		{
			"type":"file",
			"format":"json",
			"file_name":"app.log",
			"rotate":"day",
			"level":"info"
		},
		{
			"type":"file",
			"format":"json",
			"file_name":"error.log",
			"rotate":"day",
			"level":"error"
		},
		{
			"type":"console",
			"format":"simple",
			"level":"info"
		}
	]
}`

	setupLogger(cfgStr, t)

	ZLogger.Info("This is info log")
	ZLogger.Warn("This is warn log")
	ZLogger.Error("This is error log")
	ZLogger.Sugar()
}

func TestRemoveFile(t *testing.T)  {
	var cfgStr string = `{
	"common_field":{
		"app":"log-demo"
	},
	"log_dir":"./log",
	"routes":[
		{
			"type":"file",
			"format":"json",
			"file_name":"app.log",
			"rotate":"none",
			"level":"info"
		}
	]
}`

	setupLogger(cfgStr, t)
	ticker := time.NewTicker(time.Second)
	cnt := 0
	for ;;{
		cnt++
		select {
		case <-ticker.C:
			log.Println("write log", cnt)
			ZLogger.Info("log cnt ", zap.Int("cnt", cnt))
		}
	}
}

//验证点
//1.文件删除后正常写入
//2.按天滚动，按日滚动，checkExit=true|false
func TestRemove(t *testing.T)  {
	//default
	f, e := NewRotateFile("none", "log/rotate_none.log", true, nil)
	if e!= nil {
		t.Error(e)
	}
	//f, _ := os.OpenFile("log/rotate_none.log", os.O_APPEND|os.O_WRONLY, 0755)
	ticker := time.NewTicker(time.Second)
	cnt := 0
	for ;; {
		select {
		case <-ticker.C:
			msg := fmt.Sprintf("log %d\n", cnt)
			n, err := f.Write([]byte(msg))
			log.Println("write cnt", n, "err", err)
			cnt++
			if cnt%10==0 {
				break
			}
		}
	}
}

func TestDayRotate(t *testing.T)  {
	//default
	f, e := NewRotateFile("day", "log/rotate_day.log", true, nil)
	if e!= nil {
		t.Error(e)
	}

	fakeClock :=  clockwork.NewFakeClockAt(time.Now())
	clock = fakeClock

	log.Println("now", clock.Now())
	//f, _ := os.OpenFile("log/rotate_none.log", os.O_APPEND|os.O_WRONLY, 0755)
	ticker := time.NewTicker(time.Second)
	cnt := 0
	for ;; {
		select {
		case <-ticker.C:
			msg := fmt.Sprintf("log %d\n", cnt)
			_, err := f.Write([]byte(msg))
			if err!=nil {
				t.Errorf("day rotate write error:%s ", err.Error())
			}
			cnt++
			if cnt%5==0 {
				fakeClock.Advance(86400*time.Second)
				log.Println("move forward one day", fakeClock.Now())
				if _, err:=os.Stat(f.GetFileName(clock.Now())); err!=nil {
					t.Errorf("rotate failed.err:%s", err.Error())
				}
			}
		}

		if cnt%10==0 {
			break
		}
	}
}

func TestHourRotate(t *testing.T)  {
	//default
	f, e := NewRotateFile("hour", "log/rotate_hour.log", true, nil)
	if e!= nil {
		t.Error(e)
	}

	fakeClock :=  clockwork.NewFakeClockAt(time.Now())
	clock = fakeClock

	log.Println("now", clock.Now())
	//f, _ := os.OpenFile("log/rotate_none.log", os.O_APPEND|os.O_WRONLY, 0755)
	ticker := time.NewTicker(time.Second)
	cnt := 0
	for ;; {
		select {
		case <-ticker.C:
			msg := fmt.Sprintf("log %d\n", cnt)
			_, err := f.Write([]byte(msg))
			if err!=nil {
				t.Errorf("day rotate write error:%s ", err.Error())
			}
			cnt++
			if cnt%5==0 {

				fakeClock.Advance(3600*time.Second)
				log.Println("move forward one hour ", fakeClock.Now())
				if _, err:=os.Stat(f.GetFileName(clock.Now())); err!=nil {
					t.Errorf("rotate failed.err:%s", err.Error())
				}
			}
		}

		if cnt%10==0 {
			break
		}
	}
}

func TestComponent(t *testing.T)  {
	var cfgStr string = `{
	"routes":[
		{
			"type":"console",
			"level":"info"
		}
	]
}`

	setupLogger(cfgStr, t)

	logger := &CompLogger{
		Name:"iris",
		Enabled:true,
		Level: "info",
	}

	log.SetOutput(logger)
	log.Println("test compnent log")

}
func TestLevelSet(t *testing.T)  {
	var cfgStr string = `{
	"common_field":{
		"app":"log-demo"
	},
	"log_dir":"./log",
	"routes":[
		{
			"type":"console",
			"format":"json",
			"level":"info",
			"level_key":"default"
		}
	]
}`

	setupLogger(cfgStr, t)
	ZLogger.Debug("debug msg, should hide")
	ZLogger.Info("info msg, should show")

	SetLevel("default", "debug")
	ZLogger.Debug("debug msg, should show")
}

func TestSugar(t *testing.T)  {
	var cfgStr string = `{
	"routes":[
		{
			"format":"json",
			"type":"console",
			"level":"info"
		}
	]
}`

	setupLogger(cfgStr, t)
	ZLogger.Sugar().With(zap.String("key", "v")).Info("abcd")
	ctx := &Context{RefId:"1"}
	SInfo(ctx, "aaaa")
	SInfof(ctx, "abcd")
}


