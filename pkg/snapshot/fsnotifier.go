package snapshot

import (
	"log"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kaniko/pkg/util"
	"github.com/fsnotify/fsnotify"
	"github.com/karrick/godirwalk"
	"github.com/sirupsen/logrus"
)

type FsNotifier struct {
	watcher    *fsnotify.Watcher
	watchCount int
	eventsLog  map[string][]fsnotify.Op
	rootDir    string
}

func (f *FsNotifier) EventsLog() map[string][]fsnotify.Op {
	return f.eventsLog
}

func (f *FsNotifier) watchFile(path string) error {
	err := f.watcher.Add(path)
	if err != nil {
		logrus.Errorf("Failed to add watcher to path=%s, err=%s", path, err)
		return err
	}
	f.watchCount += 1
	return nil
}

func InitNotifier(rootDir string) (*FsNotifier, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Errorf("Failed to init notifier")
		return nil, err
	}
	eventsLog := make(map[string][]fsnotify.Op)

	fn := &FsNotifier{
		watcher:    w,
		rootDir:    rootDir,
		eventsLog:  eventsLog,
	}

	godirwalk.Walk(rootDir, &godirwalk.Options{
		Callback: func(path string, ent *godirwalk.Dirent) error {
			if util.CheckWhitelist(path) {
				logrus.Tracef("Not adding watches whitelisted %s", path)
				return nil
			}
			return fn.watchFile(path)
		},
		Unsorted: true,
	},
	)

	logrus.Infof("Added %d watchers", fn.watchCount)
	return fn, nil
}

func (f *FsNotifier) Close() {
	logrus.Infof("Closing watchers")
	err := f.watcher.Close()
	if err != nil {
		logrus.Warnf("Err close watcher = %s", err)
	}
}

func Start(fs *FsNotifier, done chan bool) {
	for {
		select {
		case <-done:
			logrus.Infof("Done signal")
			return
		case event, ok := <-fs.watcher.Events:
			if !ok {
				return
			}
			filePath := event.Name

			// TODO: fsnotify seem to have a bug when watching the root directory
			// the events contain double backslash instead of single one
			if strings.HasPrefix(filePath, "//") {
				filePath = strings.Replace(filePath, "//", "/", 1)
			}
			fs.eventsLog[filePath] = append(fs.eventsLog[filePath], event.Op)

			// If it's a create event, and the file path is a dir, try to create a watcher on the new dir
			if event.Op == fsnotify.Create {
				f, err := os.Stat(filePath)
				if err != nil {
					logrus.Tracef("Warning: Can not stat() a created file, skipping %s", filePath)
				} else if f.IsDir() {
					logrus.Tracef("New directory created: %s", filePath)
					_ = fs.watchFile(filePath)
				}
			}

		case err, ok := <-fs.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
