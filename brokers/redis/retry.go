package redis

import (
	"time"
)

// retryLatter 用于重试失败的操作
func retryLatter(f func(), delaySecond int) {
	fin := make(chan struct{})
	time.AfterFunc(time.Duration(delaySecond)*time.Second, func() {
		defer func() { fin <- struct{}{} }()
		f()
	})

	<-fin
}

// Retryer 重试器
type Retryer struct {
	f             func(retryTimes int) error
	retryTimes    int
	maxRetryTimes int
	err           error
	successFunc   func(retryTimes int)
	failedFunc    func(err error)
	finishFunc    func(retryTimes int, err error) bool
}

func (r *Retryer) retryFunc() {
	r.retryTimes++
	if r.err = r.f(r.retryTimes - 1); r.err != nil {
		if r.retryTimes < r.maxRetryTimes+1 {
			retryLatter(r.retryFunc, r.retryTimes)
		}
	}
}

// Retry 自动重试函数
func Retry(f func(retryTimes int) error, max int) *Retryer {
	return &Retryer{
		f:             f,
		maxRetryTimes: max,
		retryTimes:    0,
		successFunc:   func(retryTimes int) {},
		failedFunc:    func(err error) {},
		finishFunc: func(retryTimes int, err error) bool {
			return false
		},
	}
}

// Success 注册执行成功后置函数
func (r *Retryer) Success(successFunc func(retryTimes int)) *Retryer {
	r.successFunc = successFunc

	return r
}

// Failed 注册执行失败后置函数（重试最大次数后，仍然失败才调用）
func (r *Retryer) Failed(failedFunc func(err error)) *Retryer {
	r.failedFunc = failedFunc

	return r
}

// Finished 注册无论成功失败，最终执行完毕后执行的后置函数
func (r *Retryer) Finished(finishFunc func(retryTimes int, err error) bool) *Retryer {
	r.finishFunc = finishFunc

	return r
}

// Run 同步的方式运行
func (r *Retryer) Run() (int, error) {
	r.retryFunc()
	if !r.finishFunc(r.retryTimes, r.err) {
		if r.err != nil {
			r.failedFunc(r.err)
		} else {
			r.successFunc(r.retryTimes)
		}
	}

	return r.retryTimes, r.err
}

// RunAsync 异步方式运行
func (r *Retryer) RunAsync() <-chan struct{} {
	fin := make(chan struct{})

	// 由于异步执行模式下，f函数先同步执行了一次
	// 让重试次数在同步和异步模式下均表现如一
	r.retryTimes++

	// 异步执行模式下，确保第一次执行是同步的，失败时才异步去重试
	if r.err = r.f(r.retryTimes - 1); r.err == nil {
		go func() { fin <- struct{}{} }()
	} else {

		go func() {
			r.Run()
			fin <- struct{}{}
		}()
	}

	return fin
}
