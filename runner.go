package test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"

	"github.com/cjdias/flam-in-go"
)

type Runner struct {
	T     *testing.T
	App   flam.Application
	PS    flam.PubSub[string, string]
	DB    *gorm.DB
	Redis flam.RedisConnection

	wgRoutine chan struct{}
	wgRunning chan struct{}
	locker    sync.Locker

	logger Logger

	configs   []configReg
	processes []string
	arranges  []arrangeReg
	expect    error
	act       any
	asserts   []assertReg
	teardowns []teardownReg
	published []psReg
}

func NewRunner(
	app flam.Application,
	t *testing.T,
) (*Runner, error) {
	if app == nil {
		return nil, newErrNilReference("app")
	}

	return &Runner{
		T:   t,
		App: app,

		wgRoutine: make(chan struct{}),
		wgRunning: make(chan struct{}),
		locker:    &sync.Mutex{},
	}, nil
}

func (r *Runner) Close() {}

func (r *Runner) WithLogger(
	logger Logger,
) *Runner {
	r.logger = logger
	return r
}

func (r *Runner) WithConfig(
	file string,
	priority int,
) *Runner {
	r.configs = append(r.configs, configReg{file: file, priority: priority})
	return r
}

func (r *Runner) WithProcess(
	id string,
) *Runner {
	r.processes = append(r.processes, id)
	return r
}

func (r *Runner) WithArrange(
	id string,
	function any,
) *Runner {
	r.checkExecutor(function)
	r.arranges = append(r.arranges, arrangeReg{id: id, function: function})
	return r
}

func (r *Runner) WithArrangePubSub() *Runner {
	return r.WithArrange("setup pubsub", func(pubsub flam.PubSub[string, string]) {
		r.PS = pubsub
	})
}

func (r *Runner) WithArrangeSignals(
	channels []string,
) *Runner {
	_ = r.WithArrange("setup pubsub", func() {
		for _, channel := range channels {
			_ = r.PS.Subscribe("__test_runner", channel, func(channel string, message ...any) error {
				r.published = append(r.published, psReg{
					channel: channel,
					data:    message})
				return nil
			})
		}
	})
	return r.WithTeardown("teardown pubsub", func() {
		for _, channel := range channels {
			_ = r.PS.Unsubscribe("__test_runner", channel)
		}
	})
}

func (r *Runner) WithArrangeDatabase(connectionId string) *Runner {
	return r.WithArrange("begin transaction", func(factory flam.DatabaseConnectionFactory) {
		connection, _ := factory.Get(connectionId)
		r.DB = connection.Begin()
	})
}

func (r *Runner) WithArrangeRedis(connectionId string) *Runner {
	return r.WithArrange("begin redis", func(factory flam.RedisConnectionFactory) {
		connection, _ := factory.Get(connectionId)
		r.Redis = connection
	})
}

func (r *Runner) WithPanic(
	err error,
) *Runner {
	r.expect = err
	return r
}

func (r *Runner) WithAct(
	function any,
) *Runner {
	r.checkExecutor(function)
	r.act = function
	return r
}

func (r *Runner) WithAssert(
	id string,
	function any,
) *Runner {
	r.checkExecutor(function)
	r.asserts = append(r.asserts, assertReg{id: id, function: function})
	return r
}

func (r *Runner) WithAssertNoError(
	err *error,
) *Runner {
	return r.WithAssert("return no error", func() {
		if *err != nil {
			r.T.Errorf("unexpected error: %v", *err)
		}
	})
}

func (r *Runner) WithAssertDatabaseError(
	err *error,
) *Runner {
	return r.WithAssert("return database error", func() {
		switch {
		case *err == nil:
			r.T.Error("expected error, got nil")
		case (*err).Error() != "sql: database is closed":
			r.T.Errorf("unexpected 'sql: database is closed' error, get: %v", *err)
		}
	})
}

func (r *Runner) WithAssertError(
	id string,
	err *error,
	expected error,
	msg ...string,
) *Runner {
	return r.WithAssert(id, func() {
		switch {
		case *err == nil:
			r.T.Error("expected error, got nil")
		case !errors.Is(*err, expected):
			r.T.Errorf("expected %+v error, got %+v", expected, *err)
		case len(msg) > 0 && !strings.Contains((*err).Error(), msg[0]):
			r.T.Errorf("expected '%s' in error message, got %+v", msg[0], (*err).Error())
		}
	})
}

func (r *Runner) WithAssertTotal(
	expected int64,
	total *int64,
) *Runner {
	return r.WithAssert(fmt.Sprintf("return %d total number of records", expected), func() {
		if expected != *total {
			r.T.Errorf("got %d total number of records, want %d", *total, expected)
		}
	})
}

func (r *Runner) WithAssertBoolResponse(
	res *bool,
	expected bool,
) *Runner {
	return r.WithAssert(fmt.Sprintf("return %v response", expected), func() {
		if *res != expected {
			r.T.Errorf("expected %v, got %v", expected, *res)
		}
	})
}

func (r *Runner) WithAssertPublished(
	calls []string,
) *Runner {
	return r.WithAssert("assert published", func() {
		if len(r.published) != len(calls) {
			r.T.Errorf("expected %d published messages, got %d", len(calls), len(r.published))
			return
		}

		for i, c := range calls {
			if r.published[i].channel != c {
				r.T.Errorf("expected %s published message, got %s", c, r.published[i].channel)
			}
		}
	})
}

func (r *Runner) WithTeardown(
	id string,
	function any,
) *Runner {
	r.checkExecutor(function)
	r.teardowns = append(r.teardowns, teardownReg{id: id, function: function})
	return r
}

func (r *Runner) WithTeardownDatabase() *Runner {
	return r.WithTeardown("rollback transaction", func() {
		r.DB.Rollback()
	})
}

func (r *Runner) Run() error {
	var err error
	step := func(f func() error) bool {
		err = f()
		return err == nil
	}

	_ = step(r.doArrange) &&
		step(r.doAct) &&
		step(r.doAssert) &&
		step(r.doTeardown)

	assert.NoError(r.T, err, "unexpected running error")

	return err
}

func (r *Runner) doArrange() error {
	r.locker.Lock()
	err := r.App.Container().Invoke(func(
		config flam.Config,
		configSourceFactory flam.ConfigSourceFactory,
	) error {
		r.logConfigStart()

		if err := configSourceFactory.RemoveAll(); err != nil {
			return err
		}

		for _, reg := range r.configs {
			r.logConfigGetSource(reg.file)
			b, err := os.ReadFile(reg.file)
			if err != nil {
				return err
			}

			data := flam.Bag{}
			if err = yaml.NewDecoder(bytes.NewBuffer(b)).Decode(&data); err != nil {
				return err
			}

			r.logConfigRegSource(reg.file, data)
			source := &configSource{priority: reg.priority, bag: data}
			if err = configSourceFactory.Store(reg.file, source); err != nil {
				return err
			}
		}

		for _, id := range r.processes {
			_ = config.Set(fmt.Sprintf("%s.%s.active", flam.PathProcesses, id), true)
		}

		r.logConfigEnd()

		return nil
	})
	r.locker.Unlock()
	if err != nil {
		return err
	}

	r.logBootStart()
	if err := r.App.Boot(); err != nil {
		r.logBootError(err)
		return err
	}
	r.logBootEnd()

	r.logArrangeStart()
	for _, reg := range r.arranges {
		r.logArrangeRun(reg.id)
		r.locker.Lock()
		err := r.App.Container().Invoke(reg.function)
		r.locker.Unlock()
		if err != nil {
			r.logArrangeError(reg.id, err)
			return err
		}
	}
	r.logArrangeEnd()

	go func() {
		r.logAppRunStart()
		r.wgRoutine <- struct{}{}
		r.locker.Lock()
		err := r.App.Run()
		r.locker.Unlock()
		if err != nil {
			r.logAppRunError(err)
		}
		<-r.wgRunning
		r.logAppRunEnd()
	}()
	<-r.wgRoutine

	return nil
}

func (r *Runner) doAct() error {
	r.logActStart()
	r.locker.Lock()
	err := r.App.Container().Invoke(r.act)
	r.locker.Unlock()
	if err != nil {
		r.logActError(err)
	}
	r.logActEnd()
	r.wgRunning <- struct{}{}
	return err
}

func (r *Runner) doAssert() error {
	r.logAssertStart()
	for _, reg := range r.asserts {
		r.logAssertRun(reg.id)
		r.locker.Lock()
		err := r.App.Container().Invoke(reg.function)
		r.locker.Unlock()
		if err != nil {
			r.logAssertError(reg.id, err)
			return err
		}
	}
	r.logAssertEnd()
	return nil
}

func (r *Runner) doTeardown() error {
	r.logTeardownStart()
	for _, reg := range r.teardowns {
		r.logTeardownRun(reg.id)
		if err := r.App.Container().Invoke(reg.function); err != nil {
			r.logTeardownError(reg.id, err)
			return err
		}
	}
	r.logTeardownEnd()
	return nil
}

func (r *Runner) checkExecutor(
	function any,
) {
	if reflect.TypeOf(function).Kind() != reflect.Func {
		panic(errors.New("expected a function"))
	}
}

func (r *Runner) logConfigStart() {
	if r.logger != nil {
		r.logger.ConfigStart()
	}
}

func (r *Runner) logConfigGetSource(
	file string,
) {
	if r.logger != nil {
		r.logger.ConfigGetSource(file)
	}
}

func (r *Runner) logConfigRegSource(
	file string,
	data flam.Bag,
) {
	if r.logger != nil {
		r.logger.ConfigRegSource(file, data)
	}
}

func (r *Runner) logConfigEnd() {
	if r.logger != nil {
		r.logger.ConfigEnd()
	}
}

func (r *Runner) logBootStart() {
	if r.logger != nil {
		r.logger.BootStart()
	}
}

func (r *Runner) logBootError(
	err error,
) {
	if r.logger != nil {
		r.logger.BootError(err)
	}
}

func (r *Runner) logBootEnd() {
	if r.logger != nil {
		r.logger.BootEnd()
	}
}

func (r *Runner) logArrangeStart() {
	if r.logger != nil {
		r.logger.ArrangeStart()
	}
}

func (r *Runner) logArrangeRun(
	id string,
) {
	if r.logger != nil {
		r.logger.ArrangeRun(id)
	}
}

func (r *Runner) logArrangeError(
	id string,
	err error,
) {
	if r.logger != nil {
		r.logger.ArrangeError(id, err)
	}
}

func (r *Runner) logArrangeEnd() {
	if r.logger != nil {
		r.logger.ArrangeEnd()
	}
}

func (r *Runner) logAppRunStart() {
	if r.logger != nil {
		r.logger.AppRunStart()
	}
}

func (r *Runner) logAppRunError(
	err error,
) {
	if r.logger != nil {
		r.logger.AppRunError(err)
	}
}

func (r *Runner) logAppRunEnd() {
	if r.logger != nil {
		r.logger.AppRunEnd()
	}
}

func (r *Runner) logActStart() {
	if r.logger != nil {
		r.logger.ActStart()
	}
}

func (r *Runner) logActError(
	err error,
) {
	if r.logger != nil {
		r.logger.ActError(err)
	}
}

func (r *Runner) logActEnd() {
	if r.logger != nil {
		r.logger.ActEnd()
	}
}

func (r *Runner) logAssertStart() {
	if r.logger != nil {
		r.logger.AssertStart()
	}
}

func (r *Runner) logAssertRun(
	id string,
) {
	if r.logger != nil {
		r.logger.AssertRun(id)
	}
}

func (r *Runner) logAssertError(
	id string,
	err error,
) {
	if r.logger != nil {
		r.logger.AssertError(id, err)
	}
}

func (r *Runner) logAssertEnd() {
	if r.logger != nil {
		r.logger.AssertEnd()
	}
}

func (r *Runner) logTeardownStart() {
	if r.logger != nil {
		r.logger.TeardownStart()
	}
}

func (r *Runner) logTeardownRun(
	id string,
) {
	if r.logger != nil {
		r.logger.TeardownRun(id)
	}
}

func (r *Runner) logTeardownError(
	id string,
	err error,
) {
	if r.logger != nil {
		r.logger.TeardownError(id, err)
	}
}

func (r *Runner) logTeardownEnd() {
	if r.logger != nil {
		r.logger.TeardownEnd()
	}
}
