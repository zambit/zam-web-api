package providers

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/config/isc"
	"git.zam.io/wallet-backend/web-api/pkg/services/broker"
	"git.zam.io/wallet-backend/web-api/pkg/services/broker/redismq"
	"git.zam.io/wallet-backend/web-api/pkg/services/sentry"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"strings"
)

// Provide provides actual broker implementation depending on configuration
func Broker(
	container *dig.Container,
	config isc.Scheme,
	reporter sentry.IReporter,
	logger logrus.FieldLogger,
) (b broker.IBroker, e error) {
	switch {
	case strings.Contains(config.BrokerURI, "redis"):
		fallthrough
	case strings.Contains(config.BrokerURI, "rediss"):
		o, err := redis.ParseURL(config.BrokerURI)
		if err != nil {
			return
		}
		b = redismq.New(redis.NewClient(o), logger)
		b.AddMiddleware(broker.NewReportMiddleware(reporter, nil))
	case config.BrokerURI == "":
		return nil, nil
	default:
		e = fmt.Errorf("broker provider: unsopported broker url %s", config.BrokerURI)
	}

	if config.StatsEnabled {
		if rmqBroker, ok := b.(*redismq.Broker); ok {
			type statDeps struct {
				dig.In

				Group gin.IRouter `name:"root"`
			}

			e = container.Invoke(func(d statDeps) {
				d.Group.GET(config.StatsPath, func(c *gin.Context) {
					stats := rmqBroker.Connection.CollectStats(rmqBroker.Connection.GetOpenQueues())
					c.Header("Content-Type", "text/html")
					c.Writer.WriteString(stats.GetHtml("", ""))
					c.Status(200)
				})
			})
		} else {
			e = errors.New("only redis mq can provide stats")
		}
	}

	return
}
