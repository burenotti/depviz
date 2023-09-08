package app

import (
	"context"
	"depviz/internal/dependency_provider/npm"
	"depviz/internal/dependency_provider/pip"
	"depviz/internal/models"
	"depviz/internal/serializer/dot"
	"errors"
	"fmt"
	"io"
	"os"
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
	wg.Add(_defaultConcurrency)

	// tasksWg counts remaining tasks, not goroutines
	tasksWg := &sync.WaitGroup{}

	// graph with all package dependencies
	var result []models.Edge
	resMutex := &sync.Mutex{}

	// A set of visited dependencies
	visited := make(map[string]struct{})
	visMutex := &sync.Mutex{}

	// channel with fetching tasks
	taskChan := make(chan fetchTask, _defaultConcurrency)

	errChan := make(chan error)

	// may contain first caught error
	var firstErr error
	var once sync.Once

	tasksWg.Add(1)
	taskChan <- fetchTask{packageName: packageName}
	ctx2, cancel := context.WithCancel(ctx)
	for i := 0; i < _defaultConcurrency; i++ {
		go func(ctx context.Context) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-taskChan:
					if !ok {
						return
					}
					if ok := pushIfNotExist(visited, visMutex, task.packageName); !ok {
						tasksWg.Done()
						return
					}

					deps, err := a.DepsProvider.FetchPackageDeps(ctx, task.packageName)

					if err != nil {
						if errors.Is(err, context.Canceled) {
							return
						}

						once.Do(func() {
							errChan <- err
							close(errChan)
						})
						tasksWg.Done()
						return
					}
					resMutex.Lock()
					for _, dep := range deps {
						result = append(result, models.Edge{From: task.packageName, To: dep})
					}
					resMutex.Unlock()

					tasksWg.Add(len(deps))
					for _, dep := range deps {
						taskChan <- fetchTask{dep}
					}
					tasksWg.Done()
				}
			}
		}(ctx2)
	}

	done := make(chan struct{})
	go func() {
		tasksWg.Wait()
		close(done)
	}()
	select {
	case err, ok := <-errChan:
		if ok {
			firstErr = err
		}
	case <-done:
	}
	close(taskChan)
	cancel()
	wg.Wait()
	return result, firstErr
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

func Run(ctx context.Context, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	app := App{
		DepsProvider: getProviderByName(cfg.PackageManager),
		Serializer:   &dot.DotSerializer{},
	}
	return app.Run(ctx, cfg.PackageName, os.Stdout)
}

func getProviderByName(name string) DepsProvider {
	switch name {
	case "pip":
		return pip.Default()
	case "npm":
		return npm.Default()
	default:
		panic("unknown provider type: " + name)
	}
}

func pushIfNotExist(set map[string]struct{}, mu *sync.Mutex, key string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, ok := set[key]
	set[key] = struct{}{}
	return !ok
}
