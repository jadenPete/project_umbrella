package module_loader

import (
	standard_errors "errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/benbjohnson/immutable"
	"github.com/puzpuzpuz/xsync/v3"

	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/environment_variables"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/loader"
	"project_umbrella/interpreter/loader/file_loader"
	"project_umbrella/interpreter/runtime/value"
)

type ModuleLoader struct {
	cache *xsync.MapOf[string, *moduleLoaderCacheEntry]
}

func (moduleLoader *ModuleLoader) LoadFile(path_ string) value.Value {
	return moduleLoader.loadFileWithStack(path_, newModuleStack())
}

func (moduleLoader *ModuleLoader) loadFileWithStack(
	path_ string,
	moduleLoaderStack_ *moduleLoaderStack,
) value.Value {
	path_ = filepath.Clean(path_)

	if moduleLoaderStack_.Has(path_) {
		errors.RaiseError(runtime_errors.ModuleCycle(moduleLoaderStack_.ToSlice()))
	}

	entry, _ := moduleLoader.cache.LoadOrStore(path_, &moduleLoaderCacheEntry{
		result:        nil,
		computeResult: &sync.Once{},
	})

	loaderChannel := loader.NewLoaderChannel()

	go entry.computeResult.Do(
		func() {
			entry.result = file_loader.LoadFile(path_, loaderChannel)

			loaderChannel.Close()
		},
	)

	for {
		moduleName, ok := <-loaderChannel.LoadRequest

		if !ok {
			break
		}

		loaderChannel.LoadResponse <- moduleLoader.loadModuleWithStack(
			moduleName,
			moduleLoaderStack_.Add(path_),
		)
	}

	return entry.result
}

func (loader *ModuleLoader) loadModuleWithStack(
	moduleName string,
	moduleLoaderStack_ *moduleLoaderStack,
) value.Value {
	path_, ok := getModulePath(moduleName)

	if !ok {
		errors.RaiseError(runtime_errors.ModuleNotFound(moduleName))
	}

	return loader.loadFileWithStack(path_, moduleLoaderStack_)
}

func getModulePath(moduleName string) (string, bool) {
	kraitPathDirectories := strings.Split(environment_variables.KRAIT_PATH, ":")
	moduleComponents := strings.Split(moduleName, ".")

	for _, path_ := range kraitPathDirectories {
		if _, err := os.Stat(path_); standard_errors.Is(err, fs.ErrNotExist) {
			continue
		}

		currentPath := path_

		for i, component := range moduleComponents {
			if !common.IsDirectoryUnsafe(currentPath) {
				continue
			}

			currentEntries, err := os.ReadDir(currentPath)

			if err != nil {
				panic(err)
			}

			subdirectoryFound := false

			for _, entry := range currentEntries {
				if i == len(moduleComponents)-1 {
					if entry.Name() == fmt.Sprintf("%s.krait", component) {
						newPath := path.Join(currentPath, entry.Name())

						if common.IsFileUnsafe(newPath) {
							return newPath, true
						}
					}
				} else if entry.Name() == component {
					subdirectoryFound = true
					currentPath = path.Join(currentPath, entry.Name())

					break
				}
			}

			if !subdirectoryFound {
				break
			}
		}
	}

	return "", false
}

func NewModuleLoader() *ModuleLoader {
	return &ModuleLoader{
		cache: xsync.NewMapOf[string, *moduleLoaderCacheEntry](),
	}
}

type moduleLoaderCacheEntry struct {
	result        value.Value
	computeResult *sync.Once
}

type moduleLoaderStack struct {
	stackList *immutable.List[string]
	stackSet  immutable.Set[string]
}

func (stack *moduleLoaderStack) Add(moduleName string) *moduleLoaderStack {
	return &moduleLoaderStack{
		stackList: stack.stackList.Append(moduleName),
		stackSet:  stack.stackSet.Add(moduleName),
	}
}

func (stack *moduleLoaderStack) Has(moduleName string) bool {
	return stack.stackSet.Has(moduleName)
}

func (stack *moduleLoaderStack) ToSlice() []string {
	result := make([]string, 0, stack.stackList.Len())
	iterator := stack.stackList.Iterator()

	for !iterator.Done() {
		_, moduleName := iterator.Next()

		result = append(result, moduleName)
	}

	return result
}

func newModuleStack() *moduleLoaderStack {
	return &moduleLoaderStack{
		stackList: immutable.NewList[string](),
		stackSet:  immutable.NewSet[string](nil),
	}
}
