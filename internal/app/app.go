package app

import (
	"context"
	"depviz/internal/models"
	"errors"
	"fmt"
	"io"
	"sync"
)

const _defaultConcurrency = 256

type fetchTask struct {
	packageName string
}

type DepsProvider interface {
	FetchPackageDeps(ctx context.Context, packageName string) ([]string, error)
}

type Serializer interface {
	Serialize(graph []models.Edge, out io.Writer) error
}

type App struct {
	DepsProvider DepsProvider
	Serializer   Serializer
}

func New(provider DepsProvider, serializer Serializer) *App {
	return &App{
		DepsProvider: provider,
		Serializer:   serializer,
	}
}

func (a *App) GetDependencyGraph(ctx context.Context, packageName string) ([]models.Edge, error) {
	wg := &sync.WaitGroup{}
	m := &sync.Mutex{}
	mv := &sync.RWMutex{}
	var result []models.Edge
	visited := make(map[string]struct{})
	taskChan := make(chan fetchTask, _defaultConcurrency)

	wg.Add(1)
	taskChan <- fetchTask{packageName: packageName}
	ctx2, cancel := context.WithCancel(ctx)
	for i := 0; i < _defaultConcurrency; i++ {
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case task, ok := <-taskChan:
				defer wg.Done()
				if !ok {
					return
				}
				mv.Lock()
				if _, has := visited[task.packageName]; has {
					mv.Unlock()
					return
				} else {
					visited[task.packageName] = struct{}{}
				}
				mv.Unlock()

				deps, err := a.DepsProvider.FetchPackageDeps(ctx, task.packageName)
				if errors.Is(err, context.Canceled) {
					return
				} else if err != nil {
					panic(err)
				}
				wg.Add(len(deps))
				m.Lock()
				for _, dep := range deps {
					result = append(result, models.Edge{From: task.packageName, To: dep})
				}
				m.Unlock()

				for _, dep := range deps {
					taskChan <- fetchTask{dep}
				}
			}
		}(ctx2)
	}
	wg.Wait()
	cancel()
	close(taskChan)
	return result, nil
}

func (a *App) Run(ctx context.Context, packageName string, output io.Writer) error {
	graph, err := a.GetDependencyGraph(ctx, packageName)
	if err != nil {
		return fmt.Errorf("can't receive dependency graph: %w", err)
	}

	if err := a.Serializer.Serialize(graph, output); err != nil {
		return fmt.Errorf("can't write output: %w", err)
	}
	return nil
}

//func (a *App) GetDependencyGraph(ctx context.Context, packageName string) []Edge {
//	stack := []string{packageName}
//	visited := map[string]struct{}{}
//	for len(stack) > 0 {
//		top := len(stack) - 1
//		currentPackage := stack[top]
//		stack = stack[:top]
//		if _, ok := visited[currentPackage]; ok {
//			continue
//		}
//
//
//	}
//}
