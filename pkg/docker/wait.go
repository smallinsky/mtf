package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

type MultiWaitPolicy struct {
	WaitPolicies []WaitForIt
}

func (w *MultiWaitPolicy) WaitForIt(ctx context.Context, c *ContainerType) error {
	for _, wp := range w.WaitPolicies {
		if err := wp.WaitForIt(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

type WaitForIt interface {
	WaitForIt(context.Context, *ContainerType) error
}

type WaitForProcess struct {
	Process string
}

func (w *WaitForProcess) WaitForIt(ctx context.Context, c *ContainerType) error {
	return waitForHealtyAndRunning(ctx, c)
}

func (w *WaitForProcess) getHealthCheck() *HealthCheckConfig {
	return &HealthCheckConfig{
		Test:     []string{"CMD", "pgrep", w.Process},
		Interval: time.Millisecond * 100,
		Timeout:  time.Second * 1,
	}
}

type WaitWithDeley struct {
	Deley time.Duration
}

func (w *WaitWithDeley) WaitForIt(ctx context.Context, c *ContainerType) error {
	time.Sleep(w.Deley)
	return nil
}

type WaitForPort struct {
	Port int
}

func (w *WaitForPort) WaitForIt(ctx context.Context, c *ContainerType) error {
	return waitForHealtyAndRunning(ctx, c)
}

func (w *WaitForPort) getHealthCheck() *HealthCheckConfig {
	return &HealthCheckConfig{
		Test:     []string{"CMD", "nc", "-z", fmt.Sprintf("localhost:%d", w.Port)},
		Interval: time.Millisecond * 100,
		Timeout:  time.Second * 4,
	}
}

type WaitForCommand struct {
	Command string
}

func (w *WaitForCommand) WaitForIt(ctx context.Context, c *ContainerType) error {
	return waitForHealtyAndRunning(ctx, c)
}

func (w *WaitForCommand) getHealthCheck() *HealthCheckConfig {
	return &HealthCheckConfig{
		Test:     append([]string{"CMD"}, strings.Split(w.Command, " ")...),
		Interval: time.Millisecond * 100,
		Timeout:  time.Second * 3,
	}
}

func waitForHealtyAndRunning(ctx context.Context, c *ContainerType) error {
	for {
		time.Sleep(time.Millisecond * 100)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			r, err := c.GetStateV2(ctx)
			if err != nil {
				fmt.Println("got error", err)
				continue
			}
			if r.Health.Status != types.Healthy {
				continue
			}
			if r.Status != "running" {
				continue
			}
			return nil
		}
	}
}
