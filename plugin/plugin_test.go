package plugin

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"

	"github.com/robfig/cron"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyconfig"
)

type MockPluginReadyPreprocessor struct{}

func (p MockPluginReadyPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	return http.StatusOK
}

func TestPlugin(t *testing.T) {
	config := skyconfig.Configuration{}
	Convey("new plugin from non-registered transport", t, func() {
		defer unregisterAllTransports()

		createPlugin := func() {
			NewPlugin("nonexistent", "/tmp/nonexistent", []string{}, config)
		}
		So(createPlugin, ShouldPanic)
	})

	Convey("new plugin from null transport", t, func() {
		defer unregisterAllTransports()

		RegisterTransport("null", nullFactory{})

		plugin := NewPlugin("null", "/tmp/nonexistent", []string{}, config)
		So(plugin, ShouldHaveSameTypeAs, Plugin{})
		So(plugin.transport, ShouldHaveSameTypeAs, &nullTransport{})
	})

	Convey("panic unable to register timer", t, func() {
		RegisterTransport("null", nullFactory{})
		plugin := NewPlugin("null", "/tmp/nonexistent", []string{}, config)

		c := cron.New()
		panicFunc := func() {
			plugin.initTimer(c, []timerInfo{
				{"timerName", "incorrect-spec"},
			})
		}
		So(panicFunc, ShouldPanic)
	})

	Convey("init handler", t, func() {
		RegisterTransport("null", nullFactory{})
		plugin := NewPlugin("null", "/tmp/nonexistent", []string{}, config)
		Convey("init correctly with one handler", func() {
			mux := http.NewServeMux()
			plugin.initHandler(mux, router.PreprocessorRegistry{
				"plugin": MockPluginReadyPreprocessor{},
			}, []pluginHandlerInfo{
				pluginHandlerInfo{
					Name: "chima:echo",
				},
			})
			So(len(plugin.gatewayMap), ShouldEqual, 1)
			So(plugin.gatewayMap, ShouldContainKey, "/chima/echo")
		})

		Convey("init correctly with multiple handler", func() {
			mux := http.NewServeMux()
			plugin.initHandler(mux, router.PreprocessorRegistry{
				"plugin": MockPluginReadyPreprocessor{},
			}, []pluginHandlerInfo{
				pluginHandlerInfo{
					Name: "chima:echo",
				},
				pluginHandlerInfo{
					Name:    "faseng:location",
					Methods: []string{"GET"},
				},
				pluginHandlerInfo{
					Name:    "faseng:location",
					Methods: []string{"POST", "PUT"},
				},
			})
			So(len(plugin.gatewayMap), ShouldEqual, 2)
			So(plugin.gatewayMap, ShouldContainKey, "/chima/echo")
			So(plugin.gatewayMap, ShouldContainKey, "/faseng/location")
		})
	})

}
