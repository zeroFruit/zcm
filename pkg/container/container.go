package container

import (
	"io/ioutil"
	"os"
	"path"
	"simpleconman/pkg/fsutil"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Status uint32

const (
	Initial Status = iota
	Created
	Running
	Stopped
	Unknown
)

var statusValue = []string{
	"initial",
	"created",
	"running",
	"stopped",
	"unknown",
}

func (s Status) String() string {
	return statusValue[s]
}

type Id string

func (id Id) String() string {
	return string(id)
}

func genId() Id {
	return Id(strings.ReplaceAll(uuid.NewString(), "-", ""))
}

type Getter interface {
	Get(id Id) (*Instance, *Handle, error)
	List() ([]*Instance, error)
}

type Handle struct {
	id         Id
	baseDir    string
	logFile    string
	attachFile string
	exitFile   string
	getter     Getter
}

type BaseDirFn func(id Id) string
type BaseFileFn func(id Id) string

func NewHandle(getter Getter, fn BaseDirFn,
	logFileFn, attachFileFn, exitFileFn BaseFileFn) (*Handle, error) {
	id := genId()
	baseDir := fn(id)

	ok, err := fsutil.Exists(baseDir)
	if ok {
		return nil, errors.New("container directory already exists")
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot access dir")
	}
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, errors.Wrap(err, "cannot create container dir")
	}

	// TODO: is this file paths need to be validated?
	logFile := logFileFn(id)
	attachFile := attachFileFn(id)
	exitFile := exitFileFn(id)

	return &Handle{
		id:         id,
		baseDir:    fn(id),
		logFile:    logFile,
		attachFile: attachFile,
		exitFile:   exitFile,
	}, nil
}

func (h *Handle) Id() Id {
	return h.id
}

func (h *Handle) BaseDir() string {
	return h.baseDir
}

func (h *Handle) BundleDir() string {
	return path.Join(h.BaseDir(), "bundle")
}

func (h *Handle) RootfsDir() string {
	return path.Join(h.BundleDir(), "rootfs")
}

func (h *Handle) RuntimeSpecFile() string {
	return path.Join(h.BundleDir(), "config.json")
}

func (h *Handle) StateFile() string {
	return path.Join(h.BaseDir(), "state.json")
}

func (h *Handle) LogFile() string {
	return h.logFile
}

func (h *Handle) AttachFile() string {
	return h.attachFile
}

func (h *Handle) ExitFile() string {
	return h.exitFile
}

func (h *Handle) Bundle(spec []byte, rootfs string) error {
	if err := os.MkdirAll(h.BundleDir(), 0700); err != nil {
		return errors.Wrap(err, "cannot create bundle dir")
	}
	if err := fsutil.CopyDir(rootfs, h.RootfsDir()); err != nil {
		return errors.Wrap(err, "cannot copy rootfs dir")
	}
	if err := ioutil.WriteFile(h.RuntimeSpecFile(), spec, 0644); err != nil {
		return errors.Wrap(err, "cannot write OCI runtime spec file")
	}
	return nil
}

func (h *Handle) writeStatus(status Status) error {
	statefile := h.StateFile()
	// create tmp file for atmoic update
	tmpfile := statefile + ".writing"

	if err := ioutil.WriteFile(tmpfile, []byte(status.String()), 0600); err != nil {
		return errors.Wrap(err, "cannot write status to tmp file")
	}
	return os.Rename(tmpfile, statefile)
}

func (h *Handle) Created() error {
	return h.writeStatus(Created)
}

func (h *Handle) Started() error {
	return h.writeStatus(Running)
}

type Instance struct {
	Id        Id
	Pid       uint32
	CreatedAt time.Time
	StartedAt time.Time
	Status    Status
}

func (i *Instance) CanStart() bool {
	return i.Status == Created
}
