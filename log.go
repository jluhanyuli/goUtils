package goUtilsfunc

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	CtxKeyLogID     = "K_LOGID"
	CtxKeyStressTag = "K_STRESS"
	CtxKeyEnv       = "K_ENV"
	CtxKeyMethod    = "K_METHOD"
)


//使用  defer utils.FuncLog(logrus.WithFields(utils.LogCtx(ctx)), "函数名")()
func FuncLog(entry *logrus.Entry, methodList ...string) func() {
	start := time.Now()
	method := ""
	if len(methodList) == 0 {
		method = fmt.Sprintf("%v", entry.Context.Value(CtxKeyMethod))
	} else if len(methodList) > 0 {
		method = methodList[0]
	}
	entry.Infof("[func_log] %s enter", method)

	return func() {
		cost := time.Since(start).Milliseconds()
		fields := &map[string]interface{}{
			"method": method,
			"cost":   cost,
		}
		entry.WithFields(*fields).Infof("[func_log] %s exit, cost: %v", method, cost)
	}
}
