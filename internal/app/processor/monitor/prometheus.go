package monitor

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/badAkne/order-service/internal/app/processor"
	rservice "github.com/badAkne/order-service/internal/app/service"
)

type promProc struct {
	tasks []promProcTask
	chCb  chan func(ctx context.Context)
}

type promProcTask struct {
	NextCall int64
	Period   int64
	Callback func(ctx context.Context)
}

func NewPrometheusObserver(tasks ...rservice.Metered) processor.Processor {
	p := promProc{
		chCb: make(chan func(ctx context.Context), 2048),
	}

	fact := promauto.With(prometheus.DefaultRegisterer)

	for _, task := range tasks {
		tasks := task.ProvideMetrics(fact)
		for _, newTask := range tasks {
			if newTask.Callback == nil {
				continue
			}

			newTask.Period = max(newTask.Period, 1*time.Second)
			p.tasks = append(p.tasks, promProcTask{
				Callback: newTask.Callback,
			})
		}
	}

	return &p
}

func (p *promProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	nowUnix := time.Now().UnixNano()

	for i, n := 0, len(p.tasks); i < n; i++ {
		p.tasks[i].NextCall = nowUnix + p.tasks[i].Period
		p.chCb <- p.tasks[i].Callback
	}

	p.reSortTasks()

	const CheckTime = 5 * time.Second
	chTick := time.Tick(CheckTime)

	go processor.Loop(ctx, wg, p.chCb, p.execCallback)
	go processor.Loop(ctx, wg, chTick, p.checkUpcomingTasks)
}

func (p *promProc) execCallback(
	ctx context.Context, _ *sync.WaitGroup, cb func(ctx context.Context),
) {
	cb(ctx)
}

func (p *promProc) checkUpcomingTasks(
	_ context.Context, _ *sync.WaitGroup, now time.Time,
) {
	nowUnix := now.UnixNano()

	i := 0
	for n := len(p.tasks); i < n && p.tasks[i].NextCall <= nowUnix; i++ {
		p.chCb <- p.tasks[i].Callback
		p.tasks[i].NextCall += p.tasks[i].Period
	}

	if i != 0 {
		p.reSortTasks()
	}
}

func (p *promProc) reSortTasks() {
	slices.SortFunc(p.tasks, func(a, b promProcTask) int {
		switch {
		case a.NextCall < b.NextCall:
			return -1
		case a.NextCall > b.NextCall:
			return 1
		default:
			return 0
		}
	})
}
